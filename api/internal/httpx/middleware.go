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

// ClientIPExtractor holds configuration for extracting the real client IP.
type ClientIPExtractor struct {
	trustedNets []*net.IPNet
}

// NewClientIPExtractor creates an extractor with the given trusted proxy CIDRs.
// If no CIDRs are provided, only loopback addresses are trusted.
func NewClientIPExtractor(cidrs []string) (*ClientIPExtractor, error) {
	if len(cidrs) == 0 {
		// Default: trust only loopback
		_, loopback, _ := net.ParseCIDR("127.0.0.0/8")
		_, loopback6, _ := net.ParseCIDR("::1/128")
		return &ClientIPExtractor{trustedNets: []*net.IPNet{loopback, loopback6}}, nil
	}

	nets := make([]*net.IPNet, 0, len(cidrs))
	for _, cidr := range cidrs {
		_, ipnet, err := net.ParseCIDR(cidr)
		if err != nil {
			// Try parsing as single IP
			ip := net.ParseIP(cidr)
			if ip == nil {
				return nil, err
			}
			var mask net.IPMask
			if ip.To4() != nil {
				mask = net.CIDRMask(32, 32)
			} else {
				mask = net.CIDRMask(128, 128)
			}
			ipnet = &net.IPNet{IP: ip, Mask: mask}
		}
		nets = append(nets, ipnet)
	}
	return &ClientIPExtractor{trustedNets: nets}, nil
}

func (e *ClientIPExtractor) isTrustedProxy(ip string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	for _, net := range e.trustedNets {
		if net.Contains(parsed) {
			return true
		}
	}
	return false
}

func (e *ClientIPExtractor) ClientIP(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ip = r.RemoteAddr
	}

	if !e.isTrustedProxy(ip) {
		return ip
	}

	// X-Forwarded-For: client, proxy1, proxy2
	// Walk from the rightmost proxy back to find the first untrusted IP.
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		parts := strings.Split(forwarded, ",")
		for i := len(parts) - 1; i >= 0; i-- {
			candidate := strings.TrimSpace(parts[i])
			if candidate == "" {
				continue
			}
			if e.isTrustedProxy(candidate) {
				continue
			}
			// Return the first untrusted IP from the right
			return candidate
		}
		// All IPs in the chain are trusted; return the leftmost (original client)
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return strings.TrimSpace(realIP)
	}

	return ip
}

// Default extractor used when no config is provided.
var defaultExtractor, _ = NewClientIPExtractor(nil)

// ClientIP returns the client IP using the global default extractor.
func ClientIP(r *http.Request) string {
	return defaultExtractor.ClientIP(r)
}

// SetClientIPExtractor sets the global extractor used by ClientIP.
// This should be called once during server initialization.
func SetClientIPExtractor(e *ClientIPExtractor) {
	defaultExtractor = e
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

// MaxBodySize returns a middleware that limits request body size.
func MaxBodySize(limit int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, limit)
			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeaders adds common security headers to responses.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}
