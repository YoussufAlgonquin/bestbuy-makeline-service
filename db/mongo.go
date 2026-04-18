package db

import (
    "context"
    "log"
    "os"
    "time"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client
var collection *mongo.Collection

func Connect() {
    uri := os.Getenv("MONGO_URI")
    if uri == "" {
        log.Fatal("MONGO_URI environment variable not set")
    }

    var err error
    for i := 0; i < 5; i++ {
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        client, err = mongo.Connect(ctx, options.Client().ApplyURI(uri))
        cancel()
        if err == nil {
            break
        }
        log.Printf("MongoDB connection attempt %d failed: %v", i+1, err)
        time.Sleep(3 * time.Second)
    }
    if err != nil {
        log.Fatalf("Failed to connect to MongoDB: %v", err)
    }

    collection = client.Database("bestbuy").Collection("orders")
    log.Println("Connected to MongoDB")
}

func UpdateOrderStatus(orderID string, status string) error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    objID, err := primitive.ObjectIDFromHex(orderID)
    if err != nil {
        return err
    }
    filter := bson.M{"_id": objID}
    update := bson.M{"$set": bson.M{"status": status}}
    _, err = collection.UpdateOne(ctx, filter, update)
    return err
}