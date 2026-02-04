package middleware

import (
	"net/http"

	"github.com/bozz33/sublimego/flash"
)

// Flash returns a middleware that loads flash messages into the context.
func Flash(flashManager *flash.Manager) Middleware {
	if flashManager == nil {
		panic("flash.Manager is required")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			messages := flashManager.GetAndClearFromRequest(r)

			ctx := flash.WithMessages(r.Context(), messages)
			ctx = flash.WithManager(ctx, flashManager)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
