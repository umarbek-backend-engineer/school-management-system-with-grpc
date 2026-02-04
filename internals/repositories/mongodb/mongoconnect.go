package mongodb

import (
	"context"
	"log"
	"school_project_grpc/pkg/utils"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreatMongoClient() (*mongo.Client, error) {
	ctx := context.Background()

	// client, err := mongo.Connect(ctx, options.Client().ApplyURI("username:password@mongodb://localhost:27017"))
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		return nil, utils.ErrorHandler(err, "Failed to connect to DataBase")
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Failed to ping to DataBase")
	}

	log.Println("🎉 mongoDB connected successfully")
	return client, nil
}
