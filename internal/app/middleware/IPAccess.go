package middleware

import (
	"log"
	"net"
	"net/http"
)

// IPAccess creates middleware that will check that request client ip is from trusted network
func IPAccess(trustedSubnet string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var err error
			trusted := false
			if len(trustedSubnet) > 0 {
				trusted, err = isIPTrusted(r.RemoteAddr, trustedSubnet)
				if err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("Server error"))

					return
				}
			}

			if trusted {
				next.ServeHTTP(w, r)
			} else {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte("Access denied"))
			}
		})
	}
}

func isIPTrusted(addr string, subnet string) (bool, error) {
	ip, _, err := net.SplitHostPort(addr)
	if err != nil {
		return false, err
	}
	ip1 := net.ParseIP(ip)

	_, snet, err := net.ParseCIDR(subnet)
	if err != nil {
		return false, err
	}
	return snet.Contains(ip1), nil
}
