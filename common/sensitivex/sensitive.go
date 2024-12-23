package sensitivex

import (
	"os"
	"path/filepath"

	sensitive "github.com/zmexing/go-sensitive-word"
)

func NewSensitive() *sensitive.Manager {
	filter, err := sensitive.NewFilter(
		sensitive.StoreOption{Type: sensitive.StoreMemory},
		sensitive.FilterOption{Type: sensitive.FilterDfa},
	)
	if err != nil {
		panic(err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	// 加载敏感词库
	err = filter.Store.LoadDictPath(
		filepath.Join(cwd, "resources/sensitive/", "反动词库.txt"),
		filepath.Join(cwd, "resources/sensitive/", "暴恐词库.txt"),
		filepath.Join(cwd, "resources/sensitive/", "色情词库.txt"),
		filepath.Join(cwd, "resources/sensitive/", "贪腐词库.txt"),
		filepath.Join(cwd, "resources/sensitive/", "民生词库.txt"),
	)
	if err != nil {
		panic(err)
	}
	return filter
}
