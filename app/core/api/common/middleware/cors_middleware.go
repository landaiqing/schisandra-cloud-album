package middleware

import "net/http"

func CORSMiddleware() func(http.Header) {
	return func(header http.Header) {
		header.Set("Access-Control-Allow-Origin", "*")
		header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		header.Set("Access-Control-Expose-Headers", "Content-Length, Content-Type,Authorization,Accept-Language,Origin")
		header.Set("Access-Control-Allow-Credentials", "true")
	}
}
