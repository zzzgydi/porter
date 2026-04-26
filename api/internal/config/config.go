package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port int

	DatabaseURL string

	RedisAddr     string
	RedisPassword string

	RegistryInternalURL string
	RegistryPublicURL   string
	RegistryService     string
	RegistryTokenIssuer string
	RegistryTokenRealm  string
	RegistryTokenKey    string

	WebhookSecret string

	APIJWTSecret string

	BootstrapAdminEmail    string
	BootstrapAdminPassword string

	ConsoleOrigin string

	TrustedProxies []string

	DevMode bool
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	port, err := strconv.Atoi(getenv("API_PORT", "3000"))
	if err != nil {
		return nil, fmt.Errorf("invalid API_PORT: %w", err)
	}

	cfg := &Config{
		Port:                   port,
		DatabaseURL:            mustEnv("DATABASE_URL"),
		RedisAddr:              getenv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:          os.Getenv("REDIS_PASSWORD"),
		RegistryInternalURL:    getenv("REGISTRY_INTERNAL_URL", "http://localhost:5000"),
		RegistryPublicURL:      getenv("REGISTRY_PUBLIC_URL", "http://localhost:5000"),
		RegistryService:        mustEnv("REGISTRY_SERVICE"),
		RegistryTokenIssuer:    mustEnv("REGISTRY_TOKEN_ISSUER"),
		RegistryTokenRealm:     mustEnv("REGISTRY_TOKEN_REALM"),
		RegistryTokenKey:       mustEnv("REGISTRY_TOKEN_PRIVATE_KEY_PATH"),
		WebhookSecret:          mustEnv("WEBHOOK_SECRET"),
		APIJWTSecret:           mustEnv("API_JWT_SECRET"),
		BootstrapAdminEmail:    os.Getenv("BOOTSTRAP_ADMIN_EMAIL"),
		BootstrapAdminPassword: os.Getenv("BOOTSTRAP_ADMIN_PASSWORD"),
		TrustedProxies:         parseTrustedProxies(os.Getenv("TRUSTED_PROXIES")),
		DevMode:                strings.ToLower(os.Getenv("DEV_MODE")) == "true",
	}

	return cfg, nil
}

func (c *Config) GetConsoleOrigin() string {
	if c.ConsoleOrigin != "" {
		return c.ConsoleOrigin
	}
	if c.DevMode {
		return "http://localhost:4173"
	}
	return c.RegistryPublicURL
}

func (c *Config) TokenTTL() time.Duration {
	return 15 * time.Minute
}

func parseTrustedProxies(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required environment variable %s is not set", key))
	}
	return v
}
