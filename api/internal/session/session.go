package session

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Claims struct {
	UserID string `json:"sub"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Exp    int64  `json:"exp"`
}

type Manager struct {
	secret []byte
}

func NewManager(secret string) *Manager {
	return &Manager{secret: []byte(secret)}
}

func (sm *Manager) Issue(userID, email, role string, ttl time.Duration) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		Exp:    time.Now().Add(ttl).Unix(),
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	body := base64.RawURLEncoding.EncodeToString(payload)
	sig := sm.sign(header + "." + body)
	return header + "." + body + "." + sig, nil
}

func (sm *Manager) Parse(token string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}
	expectedSig := sm.sign(parts[0] + "." + parts[1])
	if !hmac.Equal([]byte(expectedSig), []byte(parts[2])) {
		return nil, fmt.Errorf("invalid signature")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decode payload: %w", err)
	}
	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("unmarshal claims: %w", err)
	}
	if time.Now().Unix() > claims.Exp {
		return nil, fmt.Errorf("token expired")
	}
	return &claims, nil
}

func (sm *Manager) sign(msg string) string {
	mac := hmac.New(sha256.New, sm.secret)
	mac.Write([]byte(msg))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
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

func FromRequest(sm *Manager, r *http.Request) (*Claims, error) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return nil, err
	}
	return sm.Parse(cookie.Value)
}
