package collection

import (
	"github.com/chenmingyong0423/go-mongox/v2"

	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
)

// MustNewCollection creates a new Collection instance with the given name.
func MustNewCollection[T any](svcCtx *svc.ServiceContext, collectionName string) *mongox.Collection[T] {
	collection := svcCtx.MongoClient.Collection(collectionName)
	return mongox.NewCollection[T](collection)
}
