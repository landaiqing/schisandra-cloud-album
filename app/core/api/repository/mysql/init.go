package mysql

import (
	"context"
	"database/sql"
	"log"
	"time"

	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/go-sql-driver/mysql"

	"schisandra-album-cloud-microservices/app/core/api/repository/mysql/ent"
)

func NewMySQL(url string) *ent.Client {
	var db *sql.DB
	db, err := sql.Open("mysql", url)

	if err != nil {
		log.Panicf("failed to connect to database: %v", err)
		return nil
	}
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Hour)
	drv := entsql.OpenDB("mysql", db)
	client := ent.NewClient(ent.Driver(drv), ent.Debug())

	if err = client.Schema.Create(context.Background()); err != nil {
		log.Panicf("failed creating model resources: %v", err)
	}
	return client
}
