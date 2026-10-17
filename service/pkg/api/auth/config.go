package auth

import "time"

type AuthConfig struct {
	AllowedOrigins    []string
	IPWhitelist       []string
	FrontendKeys      map[string]bool
	BackendKeys       map[string]bool
	SigningKeys       map[string]bool
	RateLimitMax      int
	RateLimitDuration time.Duration
}
