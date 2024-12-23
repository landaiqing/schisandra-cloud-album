package mongodb

import (
	"github.com/chenmingyong0423/go-mongox/v2"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// MustNewCollection creates a new Collection instance with the given name.
func MustNewCollection[T any](mongoClient *mongo.Database, collectionName string) *mongox.Collection[T] {
	collection := mongoClient.Collection(collectionName)
	return mongox.NewCollection[T](collection)
}
