package httpapi

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/go-chi/chi/v5/middleware"
)

// recoverer turns a panic into the same structured 500 envelope the Python
// service returned, keeping the internal message and stack in the logs rather
// than the response body.
func recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			rvr := recover()
			if rvr == nil {
				return
			}
			// http.ErrAbortHandler is a deliberate abort; re-panic so the
			// server's own handling applies and no response is written.
			if rvr == http.ErrAbortHandler {
				panic(rvr)
			}
			slog.Error("internal server error",
				"request_id", middleware.GetReqID(r.Context()),
				"method", r.Method,
				"path", r.URL.Path,
				"panic", rvr,
				"stack", string(debug.Stack()))
			writeError(w, http.StatusInternalServerError,
				"internal_server_error", "internal server error", nil)
		}()
		next.ServeHTTP(w, r)
	})
}
