package middleware

import (
	"net/http"

	"github.com/tuanvumaihuynh/outbox-pattern/pkg/correlationid"
)

func CorrelationID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			correlationID := r.Header.Get(correlationid.Header)
			if correlationID == "" {
				correlationID = correlationid.New()
				w.Header().Set(correlationid.Header, correlationID)
			}

			ctx := correlationid.NewContext(r.Context(), correlationID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
