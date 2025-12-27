package database


import (
	"log"
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"BackendFramework/internal/config"
)

var DbAuth *mongo.Database

func OpenAuth() {
	// Use the SetServerAPIOptions() method to set the Stable API version to 1
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)

	// uri := "mongodb://"+config.DB_AUTH_USERNAME+":"+config.DB_AUTH_PASSWORD+"@"+config.DB_AUTH_HOSTNAME+"/?authSource=admin"
	uri := "mongodb://" + config.DB_AUTH_HOSTNAME

	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	// Create a new client and connect to the server
	client, err := mongo.Connect(opts)

	if err != nil {
		log.Fatalf("Failed to connect to DB AUTH %v", err)
	}
	
	if err := client.Ping(context.TODO(), nil); err != nil {
		log.Fatalf("Failed to ping DB AUTH %v", err)
	}

	DbAuth = client.Database(config.DB_AUTH_DBNAME)
}