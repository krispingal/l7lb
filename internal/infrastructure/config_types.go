package infrastructure

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

// RateLimiter defines the structure for rate limiter configuration
type RateLimiter struct {
	Type   string `mapstructure:"type"`   // e.g. "none", "sliding_window", "token_bucket"
	Limit  int    `mapstructure:"limit"`  // request limit for the time window/bucket
	Window string `mapstructure:"window"` // only for window-based rate limiters
}

// LoadBalancer holds the load balancer address
type LoadBalancer struct {
	Address  string `mapstructure:"address"`
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
}

type Config struct {
	Routes       []Route      `mapstructure:"routes"`
	RateLimiter  RateLimiter  `mapstructure:"rateLimiter"`
	LoadBalancer LoadBalancer `mapstructure:"loadbalancer"`
}
