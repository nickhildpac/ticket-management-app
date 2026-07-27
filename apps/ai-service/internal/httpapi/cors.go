package httpapi

import (
	"net/http"
	"slices"
)

// configured origins with credentials, any method, any header.
//
// In the default topology the browser talks to the ticket-service, which
// proxies ingest server-side, so this only matters when the ai-service is
// called directly from a browser.
func corsMiddleware(origins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && slices.Contains(origins, origin) {
				h := w.Header()
				h.Set("Access-Control-Allow-Origin", origin)
				h.Set("Access-Control-Allow-Credentials", "true")
				h.Set("Vary", "Origin")
				if r.Method == http.MethodOptions {
					h.Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
					// Echo the requested headers rather than hard-coding a list,
					// matching FastAPI's allow_headers=["*"].
					if req := r.Header.Get("Access-Control-Request-Headers"); req != "" {
						h.Set("Access-Control-Allow-Headers", req)
					}
					w.WriteHeader(http.StatusNoContent)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
