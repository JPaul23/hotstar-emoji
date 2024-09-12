package controllers

import (
	"api/database"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// reactions request
type ReactionRequest struct {
	EmojiID string `json:"emojiId" binding:"required"`
}

type Emoji struct {
	ID          primitive.ObjectID `bson:"_id" json:"id"`
	Emoji       string             `bson:"symbol"`
	Description string             `bson:"title"`
	Keywords    string             `bson:"keywords"`
	Code        string             `bson:"code"`
}

var collection, _ = database.InitDb()

// var messageChannel = make(chan []byte, 1000) // Buffer size 1000

// func HandleKafkaProducerBg() {
// 	// Producer goroutine to flush data periodically to Kafka
// 	go func() {
// 		for {
// 			select {
// 			case message := <-messageChannel:
// 				messageStr := string(message)
// 				log.Println("Kafka message ====> :", messageStr)
// 				// TODOSend message to Kafka
// 				err := kafka.KafkaMessageProducer(producer, []byte("emoji-reaction"), message)
// 				if err != nil {
// 					log.Println("Error sending message to Kafka:", err)
// 				}
// 				// TODO: handle the flushing
// 				case <-time.After(500 * time.Millisecond):
// 				    // Flush data to Kafka every 500ms
// 				    // Implement the logic to flush data to Kafka
// 			}
// 		}
// 	}()
// }

func GetEmojis(c *gin.Context) {
	emojis := []Emoji{} // Assuming Emoji is the struct representing your emojis
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}
	limit := c.Query("limit")
	fmt.Println(limit)
	if limit == "" {
		limit = "5"
	}

	limitNum, err := strconv.Atoi(limit)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "Invalid limit parameter",
		})
		return
	}

	cursor, err := collection.Find(context.Background(), bson.D{{}}, options.Find().SetLimit(int64(limitNum)))
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Failed to fetch emojis",
		})
		return
	}

	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var emoji Emoji
		err := cursor.Decode(&emoji)
		if err != nil {
			log.Fatal(err)
		}
		emojis = append(emojis, emoji)
	}

	c.JSON(200, gin.H{
		"emojis":  emojis,
		"Message": "Emoji successfully reached",
	})
}

func EmojiReaction(c *gin.Context, messageChannel chan kafka.Message) {
	// var workerLimit = 10
	// Declare a channel to control concurrency

	var reaction ReactionRequest
	err := c.ShouldBindJSON(&reaction)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "Invalid request payload",
		})
		return
	}

	objectId, err := primitive.ObjectIDFromHex(reaction.EmojiID)
	if err != nil {
		c.JSON(404, gin.H{
			"Message": "Emoji Not found!",
		})
		return
	}

	filter := bson.M{"_id": objectId}

	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Failed to fetch emoji",
		})
		return
	}

	defer cursor.Close(context.Background())
	// Initialize Kafka producer
	// producer := kafka.KafkaProducerSetup()
	// defer producer.Close()

	var emoji Emoji
	for cursor.Next(context.Background()) {
		err := cursor.Decode(&emoji)
		if err != nil {
			log.Fatal(err)
		}

		// // Wait for all goroutines to finish
		// for i := 0; i < workerLimit; i++ {
		// 	concurrency <- struct{}{}
		// }
	}
	emojiJSON, err := json.Marshal(emoji)
	if err != nil {
		log.Println("Error marshalling data to JSON:", err)
	}

	// Create a Kafka message
	message := kafka.Message{
		Key:   []byte("emoji-reaction"),
		Value: emojiJSON,
	}

	// Add message to the channel for production
	messageChannel <- message

	c.JSON(200, gin.H{
		"Message": "Reaction successfully logged",
	})
}
