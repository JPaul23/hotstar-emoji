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
		panic("Error loading .env file")
	}
	// Get value from .env
	MONGO_URI := os.Getenv("MONGO_URI")
	DATABASE := os.Getenv("DATABASE")
	COLLECTION := os.Getenv("COLLECTION")

	// Connect to the database.
	clientOptions := options.Client().ApplyURI(MONGO_URI)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection.
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	collection := client.Database(DATABASE).Collection(COLLECTION)

	fmt.Println("Connected to db")
	return collection, nil
}
