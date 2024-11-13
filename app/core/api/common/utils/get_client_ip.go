package utils

import (
	"net"
	"net/http"
	"strings"
)

// GetClientIP returns the client IP address from the request.
func GetClientIP(r *http.Request) string {
	xForwardedFor := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	if xForwardedFor != "" {
		return strings.Split(xForwardedFor, ",")[0]
	}

	ip := strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	if ip != "" {
		return ip
	}

	ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err != nil {
		return ""
	}
	return ip
}
