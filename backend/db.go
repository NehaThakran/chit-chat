package main

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MessagesCollection *mongo.Collection

func InitDB() {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017/chatdb"
	}

	 log.Printf("Connecting to MongoDB at: %s\n", mongoURI)
	// Set client options
	clientOptions := options.Client().ApplyURI(mongoURI)

	//create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to MongoDB!")
	
	// Get a handle for the "messages" table in the "chatdb" database
	MessagesCollection = client.Database("chatdb").Collection("messages")
}