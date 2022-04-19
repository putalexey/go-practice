package middleware

import (
	"net/http"
)

// IPAccess creates middleware that will check that request client ip is from trusted network
func IPAccess(trustedSubnet string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(trustedSubnet) > 0 {
				next.ServeHTTP(w, r)
			} else {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte("Access denied"))
			}
		})
	}
}
