package middleware

import "net/http"

func CORSMiddleware() func(http.Header) {
	return func(header http.Header) {
		header.Set("Access-Control-Allow-Origin", "*")
		header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		header.Set("Access-Control-Expose-Headers", "Content-Length, Content-Type")
		header.Set("Access-Control-Allow-Headers", "Content-Type,Authorization,Accept-Language,Origin,X-Content-Security,X-UID")
		header.Set("Access-Control-Allow-Credentials", "true")
	}
}
