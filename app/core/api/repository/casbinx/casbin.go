package casbinx

import (
	"os"
	"path/filepath"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

// NewCasbin creates a new casbinx enforcer with a mysql adapter and loads the policy from the file system.
func NewCasbin(engine *gorm.DB) *casbin.SyncedCachedEnforcer {
	a, err := gormadapter.NewAdapterByDBUseTableName(engine, "sca_auth_", "permission_rule")
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
	e, err := casbin.NewSyncedCachedEnforcer(modelFile, a)
	if err != nil {
		panic(err)
	}
	e.EnableCache(true)
	err = e.LoadPolicy()
	if err != nil {
		panic(err)
	}
	return e
}
