package middleware

import (
	"context"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/slchris/wg-mgt/internal/pkg/response"
	"github.com/slchris/wg-mgt/internal/service"
)

type contextKey string

const (
	// UserContextKey is the context key for user claims.
	UserContextKey contextKey = "user"
)

// Logger logs HTTP requests.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

// Recovery recovers from panics.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %v", err)
				response.InternalError(w, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// LocalOnly restricts access to localhost.
func LocalOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			host = r.RemoteAddr
		}

		ip := net.ParseIP(host)
		if ip == nil || (!ip.IsLoopback() && !isLocalhost(host)) {
			response.Forbidden(w, "access denied")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// isLocalhost checks if the host is localhost.
func isLocalhost(host string) bool {
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}

// Auth validates JWT tokens.
func Auth(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Unauthorized(w, "missing authorization header")
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				response.Unauthorized(w, "invalid authorization header")
				return
			}

			claims, err := authService.ValidateToken(parts[1])
			if err != nil {
				response.Unauthorized(w, "invalid token")
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserFromContext retrieves user claims from context.
func GetUserFromContext(ctx context.Context) *service.Claims {
	if claims, ok := ctx.Value(UserContextKey).(*service.Claims); ok {
		return claims
	}
	return nil
}
