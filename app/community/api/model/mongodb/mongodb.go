package mongodb

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

// NewMongoDB initializes the MongoDB connection and returns the database object
func NewMongoDB(uri string, username string, password string, authSource string, database string) *mongo.Database {
	client, err := mongo.Connect(options.Client().ApplyURI(uri).SetAuth(options.Credential{
		Username:   username,
		Password:   password,
		AuthSource: authSource,
	}))
	if err != nil {
		panic(err)
	}
	err = client.Ping(context.Background(), readpref.Primary())
	if err != nil {
		panic(err)
	}
	db := client.Database(database)
	return db
}
