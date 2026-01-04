package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// OfflineTenantMiddleware injects the default tenant ID into the request context
// This middleware should be used in offline mode where there's only one tenant
func OfflineTenantMiddleware(defaultTenantID uuid.UUID) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), TenantIDKey, defaultTenantID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

