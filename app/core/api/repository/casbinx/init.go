package casbinx

import (
	"os"
	"path/filepath"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"

	"schisandra-album-cloud-microservices/app/core/api/repository/casbinx/adapter"
)

// NewCasbin creates a new casbinx enforcer with a mysql adapter and loads the policy from the file system.
func NewCasbin(dataSourceName string) *casbin.CachedEnforcer {
	a, err := adapter.NewAdapter("mysql", dataSourceName)
	if err != nil {
		panic(err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	path := filepath.Join(cwd, "etc", "rbac_model.conf")
	modelFile, err := model.NewModelFromFile(path)
	if err != nil {
		panic(err)
	}
	e, err := casbin.NewCachedEnforcer(modelFile, a)
	if err != nil {
		panic(err)
	}
	e.EnableCache(true)
	e.SetExpireTime(60 * 60)
	err = e.LoadPolicy()
	if err != nil {
		panic(err)
	}
	return e
}
