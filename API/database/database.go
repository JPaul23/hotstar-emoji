package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitDb() (*mongo.Collection, error) {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Error loading .env file:", err)

	}
	// Get value from .env
	MONGO_URI := os.Getenv("MONGO_URI")
	DATABASE := os.Getenv("DATABASE")
	COLLECTION := os.Getenv("COLLECTION")

	log.Println("MONGO_URI:", MONGO_URI)
	log.Println("DATABASE:", DATABASE)
	log.Println("COLLECTION:", COLLECTION)
	// Check if environment variables are set
	if MONGO_URI == "" || DATABASE == "" || COLLECTION == "" {
		return nil, fmt.Errorf("environment variables MONGO_URI, DATABASE, and COLLECTION must be set")
	}

	// Connect to the database.
	clientOptions := options.Client().ApplyURI(MONGO_URI)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal("Mongo.Connect error ====> ", err)
	}

	// Check the connection.
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal("client.Ping ERROR ======>", err)
	}
	collection := client.Database(DATABASE).Collection(COLLECTION)

	fmt.Println("Connected to db")
	return collection, nil
}
