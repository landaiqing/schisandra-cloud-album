package middleware

import "net/http"

type I18nMiddleware struct {
}

func NewI18nMiddleware() *I18nMiddleware {
	return &I18nMiddleware{}
}

func (m *I18nMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO generate middleware implement function, delete after code implementation

		// Passthrough to next handler if need
		next(w, r)
	}
}
