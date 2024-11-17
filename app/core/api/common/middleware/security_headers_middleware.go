package middleware

import "net/http"

func SecurityHeadersMiddleware(r *http.Request) {
	r.Header.Set("X-Frame-Options", "DENY")
	r.Header.Set("Content-Security-Policy", "default-src 'self'; connect-src *; font-src *; script-src-elem * 'unsafe-inline'; img-src * data:; style-src * 'unsafe-inline';")
	r.Header.Set("X-XSS-Protection", "1; mode=block")
	r.Header.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
	r.Header.Set("Referrer-Policy", "strict-origin")
	r.Header.Set("X-Content-Type-Options", "nosniff")
	r.Header.Set("Permissions-Policy", "geolocation=(),midi=(),sync-xhr=(),microphone=(),camera=(),magnetometer=(),gyroscope=(),fullscreen=(self),payment=()")
}
