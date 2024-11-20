package casbinx

import (
	"os"
	"path/filepath"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	xormadapter "github.com/casbin/xorm-adapter/v3"
	_ "github.com/go-sql-driver/mysql"
	"xorm.io/xorm"
)

// NewCasbin creates a new casbinx enforcer with a mysql adapter and loads the policy from the file system.
func NewCasbin(engine *xorm.Engine) *casbin.CachedEnforcer {
	a, err := xormadapter.NewAdapterByEngineWithTableName(engine, "permission_rule", "sca_auth_")
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
