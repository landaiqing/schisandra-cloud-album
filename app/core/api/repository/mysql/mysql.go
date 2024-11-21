package mysql

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"schisandra-album-cloud-microservices/app/core/api/repository/mysql/query"
)

func NewMySQL(url string, maxOpenConn int, maxIdleConn int) (*gorm.DB, *query.Query) {
	db, err := gorm.Open(mysql.Open(url), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             time.Second, // 慢sql日志
				LogLevel:                  logger.Info, // 级别
				Colorful:                  true,        // 颜色
				IgnoreRecordNotFoundError: true,        // 忽略RecordNotFoundError
				ParameterizedQueries:      true,        // 格式化SQL语句
			}),
	})
	if err != nil {
		panic(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}
	sqlDB.SetMaxOpenConns(maxOpenConn)
	sqlDB.SetMaxIdleConns(maxIdleConn)
	useDB := query.Use(db)
	return db, useDB
}
