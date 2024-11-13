package ip2region

import (
	"os"
	"path/filepath"

	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
)

// NewIP2Region creates a new IP2Region searcher instance.
func NewIP2Region() *xdb.Searcher {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	dbPath := filepath.Join(cwd, "resources/ip2region", "ip2region.xdb")
	cBuff, err := xdb.LoadContentFromFile(dbPath)
	if err != nil {
		panic(err)
	}
	searcher, err := xdb.NewWithBuffer(cBuff)
	if err != nil {
		panic(err)
		return nil
	}
	return searcher
}
