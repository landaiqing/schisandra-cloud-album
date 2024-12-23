package middleware

import (
	"net/http"
	"os"
	"path/filepath"
	"schisandra-album-cloud-microservices/common/i18n"

	"golang.org/x/text/language"
)

func I18nMiddleware(next http.HandlerFunc) http.HandlerFunc {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	zhPath := filepath.Join(cwd, "/resources/language/", "active.zh.toml")
	enPath := filepath.Join(cwd, "/resources/language/", "active.en.toml")
	return i18n.NewI18nMiddleware([]language.Tag{
		language.English,
		language.Chinese,
	}, []string{enPath, zhPath}).Handle(next)
}
