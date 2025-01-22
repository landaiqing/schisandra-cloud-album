package mysql

import (
	"log"
	"os"
	"schisandra-album-cloud-microservices/app/aisvc/model/mysql/model"
	"schisandra-album-cloud-microservices/app/aisvc/model/mysql/query"
	"time"

	"github.com/asjdf/gorm-cache/cache"
	"github.com/asjdf/gorm-cache/config"
	"github.com/asjdf/gorm-cache/storage"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewMySQL(url string, maxOpenConn int, maxIdleConn int, client *redis.Client) (*gorm.DB, *query.Query) {
	db, err := gorm.Open(mysql.Open(url), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             time.Second,  // 慢sql日志
				LogLevel:                  logger.Error, // 级别
				Colorful:                  true,         // 颜色
				IgnoreRecordNotFoundError: true,         // 忽略RecordNotFoundError
				ParameterizedQueries:      true,         // 格式化SQL语句
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
	// migrate
	Migrate(db)
	// cache
	gormCache, err := cache.NewGorm2Cache(&config.CacheConfig{
		CacheLevel: config.CacheLevelAll,
		CacheStorage: storage.NewRedis(&storage.RedisStoreConfig{
			KeyPrefix: "cache",
			Client:    client,
		}),
		InvalidateWhenUpdate:           true,  // when you create/update/delete objects, invalidate cache
		CacheTTL:                       10000, // 5000 ms
		CacheMaxItemCnt:                0,     // if length of objects retrieved one single time
		AsyncWrite:                     true,  // async write to cache
		DebugMode:                      false,
		DisableCachePenetrationProtect: true, // disable cache penetration protect
	})
	if err != nil {
		panic(err)
	}
	err = db.Use(gormCache)
	if err != nil {
		panic(err)
	}

	return db, useDB
}

func Migrate(db *gorm.DB) {
	err := db.AutoMigrate(
		&model.ScaStorageFace{})
	if err != nil {
		panic(err)
	}
}
