package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	MongoString       = os.Getenv("MONGOSTRING")
	DBName            = "senyum_pelanggan"
	RiwayatCollection = "riwayat_emosi"
	UserCollection    = "user"
	DB                *mongo.Database
)

// ConnectDB initializes the database connection and stores it in DB
func ConnectDB() {
	if MongoString == "" {
		MongoString = "mongodb://localhost:27017" // fallback default
	}

	client, err := mongo.NewClient(options.Client().ApplyURI(MongoString))
	if err != nil {
		log.Fatalf("Gagal buat MongoDB client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatalf("Gagal konek ke MongoDB: %v", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Gagal ping MongoDB: %v", err)
	}

	DB = client.Database(DBName)
	fmt.Println("Berhasil konek ke MongoDB!")
}

// GetCollection returns a collection instance by name
func GetCollection(collectionName string) *mongo.Collection {
	return DB.Collection(collectionName)
}
