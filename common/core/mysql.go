package core

import (
	"context"
	"database/sql"
	"log"

	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/go-sql-driver/mysql"

	"schisandra-album-cloud-microservices/common/ent"
)

func InitMySQL(url string) *ent.Client {
	var db *sql.DB
	db, err := sql.Open("mysql", url)

	if err != nil {
		return nil
	}
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(50)
	drv := entsql.OpenDB("mysql", db)
	client := ent.NewClient(ent.Driver(drv))

	defer client.Close()

	if err = client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}
	return client
}
