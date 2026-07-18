package session

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

type Claims struct {
	UserID string `json:"sub"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

type Manager struct {
	secret []byte
}

func NewManager(secret string) *Manager {
	return &Manager{secret: []byte(secret)}
}

func (sm *Manager) Issue(userID, email, role string, ttl time.Duration) (string, error) {
	tok := jwt.New()
	_ = tok.Set(jwt.SubjectKey, userID)
	_ = tok.Set("email", email)
	_ = tok.Set("role", role)
	_ = tok.Set(jwt.IssuedAtKey, time.Now())
	_ = tok.Set(jwt.NotBeforeKey, time.Now().Add(-10*time.Second))
	_ = tok.Set(jwt.ExpirationKey, time.Now().Add(ttl))

	signed, err := jwt.Sign(tok, jwt.WithKey(jwa.HS256, sm.secret))
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return string(signed), nil
}

func (sm *Manager) Parse(token string) (*Claims, error) {
	tok, err := jwt.ParseString(token,
		jwt.WithKey(jwa.HS256, sm.secret),
		jwt.WithValidate(true),
		jwt.WithAcceptableSkew(5*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims := &Claims{}
	claims.UserID = tok.Subject()
	if v, ok := tok.Get("email"); ok {
		claims.Email, _ = v.(string)
	}
	if v, ok := tok.Get("role"); ok {
		claims.Role, _ = v.(string)
	}
	return claims, nil
}

func (sm *Manager) Cookie(token string, secure bool, ttl time.Duration) *http.Cookie {
	sameSite := http.SameSiteLaxMode
	if secure {
		sameSite = http.SameSiteStrictMode
	}
	return &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		MaxAge:   int(ttl.Seconds()),
	}
}

// ErrSessionInvalid is returned when a session was present but is no longer
// valid (e.g. the user was deleted after the token was issued).
var ErrSessionInvalid = errors.New("session invalid")

type claimsContextKey struct{}

// ContextWithClaims stores pre-validated claims on the context, allowing
// middleware to inject claims refreshed from the database. Storing a nil
// Claims pointer marks the session as invalid.
func ContextWithClaims(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, claimsContextKey{}, claims)
}

func FromRequest(sm *Manager, r *http.Request) (*Claims, error) {
	// Middleware may have already validated and refreshed the claims.
	if claims, ok := r.Context().Value(claimsContextKey{}).(*Claims); ok {
		if claims == nil {
			return nil, ErrSessionInvalid
		}
		return claims, nil
	}
	cookie, err := r.Cookie("session")
	if err != nil {
		return nil, err
	}
	return sm.Parse(cookie.Value)
}
