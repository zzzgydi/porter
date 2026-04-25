package httpx

import (
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func ClientIP(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ip = r.RemoteAddr
	}

	if isTrustedProxy(ip) {
		realIP := r.Header.Get("X-Real-IP")
		if realIP != "" {
			ip = realIP
		} else {
			forwarded := r.Header.Get("X-Forwarded-For")
			if forwarded != "" {
				// X-Forwarded-For can contain multiple IPs: client, proxy1, proxy2
				// Take the first one (the actual client)
				if idx := strings.Index(forwarded, ","); idx != -1 {
					ip = strings.TrimSpace(forwarded[:idx])
				} else {
					ip = strings.TrimSpace(forwarded)
				}
			}
		}
	}
	return ip
}

func isTrustedProxy(ip string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	return parsed.IsLoopback() || parsed.IsPrivate()
}

func RequestID(next http.Handler) http.Handler {
	return middleware.RequestID(next)
}

func Logger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
			logger.Info("http",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"duration", time.Since(start).String(),
				"remote", ClientIP(r),
			)
		})
	}
}

func Recoverer(next http.Handler) http.Handler {
	return middleware.Recoverer(next)
}

func CORS(origin string) func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   []string{origin},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})
}
