package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"api/controllers"

	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

var producer *kafka.Writer

// var messageChannel = make(chan []byte, 1000) // Buffer size 1000

func logRequestMiddleware() gin.HandlerFunc {
	logger := zap.NewExample()
	defer logger.Sync() // flushes buffer, if any

	return func(ctx *gin.Context) {
		logger.Info("incoming request",
			zap.String("path", ctx.Request.URL.Path),
			zap.String("method", ctx.Request.Method),
		)
		ctx.Next()
	}
}

// set up producer
func KafkaProducerSetup() *kafka.Writer {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file KafkaProducerSetup ====> ", err)
	}
	kafkaHost := os.Getenv("KAKFA_HOST")
	kafkaPort := os.Getenv("KAFKA_PORT")
	emojiKafkaTopic := os.Getenv("KAFKA_TOPIC")

	writer := kafka.NewWriter(kafka.WriterConfig{
		// Brokers: []string{"%s:%s", kafkaHost, kafkaPort},
		Brokers:  []string{fmt.Sprintf("%s:%s", kafkaHost, kafkaPort)},
		Topic:    emojiKafkaTopic,
		Balancer: &kafka.LeastBytes{},
	})

	return writer
}

func handleKafkaProducer(messageChannel chan kafka.Message) {
	go func() {
		err := godotenv.Load()
		if err != nil {
			log.Println("Error loading .env file handleKafkaProducer ====> ", err)
		}
		kafkaFlushTime := os.Getenv("KAFKA_FLUSH_TIME")
		if kafkaFlushTime == "" {
			log.Println("Kafka flush time is not available")
		}
		flushTime, err := strconv.Atoi(kafkaFlushTime)
		if err != nil {
			log.Println("Error converting Kafka flush time to integer")
		}
		for {
			select {
			case message := <-messageChannel:
				err := producer.WriteMessages(context.Background(), message)
				if err != nil {
					log.Println("Error sending message to Kafka:", err)
				}
			case <-time.After(time.Duration(flushTime) * time.Millisecond):
				// Flush data to Kafka every 500ms
				log.Println("Flushing data =====:")
			}
		}
	}()
}

func main() {

	producer = KafkaProducerSetup()
	defer producer.Close()
	messageChannel := make(chan kafka.Message, 1000) // Buffer size 1000

	// Set up signal handling to gracefully shut down the producer
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigchan
		producer.Close()
		os.Exit(0)
	}()

	// Set up the router
	r := setupRouter(messageChannel)

	go handleKafkaProducer(messageChannel)

	// Start the server
	apiPort := os.Getenv("API_PORT")
	r.Run(":" + apiPort)
}

// setupRouter sets up the router and adds the routes.
func setupRouter(messageChannel chan kafka.Message) *gin.Engine {
	// Create a new router
	router := gin.Default()
	router.ForwardedByClientIP = true
	router.SetTrustedProxies([]string{"127.0.0.1"})

	router.Use(logRequestMiddleware())
	// Add a welcome route
	router.GET("/", func(c *gin.Context) {
		c.String(200, "Hello from API")
	})
	// Create a new group for the API
	api := router.Group("/api/v1/hotstar")
	{
		// Create a new group for the account routes
		public := api.Group("/emoji")
		{
			// Add the login route
			public.GET("/", controllers.GetEmojis)
			// Add the signup route
			public.POST("/react", func(c *gin.Context) {
				controllers.EmojiReaction(c, messageChannel)
			})
		}

	}
	// Return the router
	return router
}
