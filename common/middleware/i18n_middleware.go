package middleware

import (
	"net/http"

	"golang.org/x/text/language"

	"schisandra-album-cloud-microservices/common/i18n"
)

func I18nMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return i18n.NewI18nMiddleware([]language.Tag{
		language.English,
		language.Chinese,
	}, []string{"../../resources/language/active.en.toml", "../../resources/language/active.zh.toml"}).Handle(next)
}
