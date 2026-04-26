package registrytoken

import (
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

type AccessEntry struct {
	Type    string   `json:"type"`
	Name    string   `json:"name"`
	Actions []string `json:"actions"`
}

type Signer struct {
	key    *rsa.PrivateKey
	jwkKey jwk.Key
	kid    string
	issuer string
}

func NewSigner(privateKeyPath string) (*Signer, error) {
	info, err := os.Stat(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("stat private key: %w", err)
	}
	mode := info.Mode().Perm()
	if mode != 0600 {
		return nil, fmt.Errorf("private key %s has permissions %04o; expected 0600", privateKeyPath, mode)
	}

	data, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("read private key: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// Try PKCS8
		keyInterface, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err2 != nil {
			return nil, fmt.Errorf("parse private key: %w", err)
		}
		var ok bool
		key, ok = keyInterface.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("not an RSA private key")
		}
	}

	kid, err := deriveKid(key)
	if err != nil {
		return nil, fmt.Errorf("derive kid: %w", err)
	}

	jwkKey, err := jwk.FromRaw(key)
	if err != nil {
		return nil, fmt.Errorf("jwk from raw: %w", err)
	}
	_ = jwkKey.Set(jwk.KeyIDKey, kid)

	return &Signer{
		key:    key,
		jwkKey: jwkKey,
		kid:    kid,
	}, nil
}

func (s *Signer) SetIssuer(issuer string) {
	s.issuer = issuer
}

func (s *Signer) Sign(subject, audience string, access []AccessEntry, ttl time.Duration) (string, error) {
	tok := jwt.New()
	_ = tok.Set(jwt.IssuerKey, s.issuer)
	_ = tok.Set(jwt.SubjectKey, subject)
	_ = tok.Set(jwt.AudienceKey, audience)
	_ = tok.Set(jwt.IssuedAtKey, time.Now())
	_ = tok.Set(jwt.NotBeforeKey, time.Now().Add(-10*time.Second))
	_ = tok.Set(jwt.ExpirationKey, time.Now().Add(ttl))
	_ = tok.Set(jwt.JwtIDKey, uuid.New().String())
	_ = tok.Set("access", access)

	signed, err := jwt.Sign(tok, jwt.WithKey(jwa.RS256, s.jwkKey))
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return string(signed), nil
}

func deriveKid(key *rsa.PrivateKey) (string, error) {
	// RFC 7638 JWK thumbprint for RSA: SHA256 of {"e":"...","kty":"RSA","n":"..."}
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.PublicKey.E)).Bytes())
	n := base64.RawURLEncoding.EncodeToString(key.PublicKey.N.Bytes())
	jwkJSON, err := json.Marshal(map[string]string{
		"e":   e,
		"kty": "RSA",
		"n":   n,
	})
	if err != nil {
		return "", fmt.Errorf("jwk marshal failed: %w", err)
	}
	hash := sha256.Sum256(jwkJSON)
	return base64.RawURLEncoding.EncodeToString(hash[:]), nil
}
