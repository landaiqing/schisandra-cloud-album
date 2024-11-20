package mysql

import (
	_ "github.com/go-sql-driver/mysql"
	"xorm.io/xorm"
	"xorm.io/xorm/caches"
	"xorm.io/xorm/log"
)

func NewMySQL(url string, maxOpenConn int, maxIdleConn int) *xorm.Engine {
	engine, err := xorm.NewEngine("mysql", url)
	if err != nil {
		panic(err)
	}
	err = engine.Ping()
	if err != nil {
		panic(err)
	}
	err = SyncDatabase(engine)
	if err != nil {
		panic(err)
	}
	engine.SetMaxOpenConns(maxOpenConn)
	engine.SetMaxIdleConns(maxIdleConn)

	cacher := caches.NewLRUCacher(caches.NewMemoryStore(), 1000)
	engine.SetDefaultCacher(cacher)

	engine.ShowSQL(true)
	engine.Logger().SetLevel(log.LOG_DEBUG)
	return engine
}
