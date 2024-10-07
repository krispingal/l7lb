package infrastructure

import "time"

// Route holds the backends for each route
type Route struct {
	Path     string
	Backends []Backend `mapstructure:"backends"`
}

// Backend holds the individual backend server configuration
type Backend struct {
	URL    string `mapstructure:"url"`
	Health string `mapstructure:"health"`
}

// Fixed window rate limiter config
type FixedWindowRateLimiter struct {
	Limit  int           `mapstructure:"limit"`
	Window time.Duration `mapstructure:"window"`
}

// LoadBalancer holds the load balancer address
type LoadBalancer struct {
	Address string `mapstructure:"address"`
}

type Config struct {
	Routes       []Route                `mapstructure:"routes"`
	RateLimiter  FixedWindowRateLimiter `mapstructure:"rateLimiter"`
	LoadBalancer LoadBalancer           `mapstructure:"loadbalancer"`
}
