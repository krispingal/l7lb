package domain

type Backend struct {
	URL    string
	Health string // holds the endpoint for healthchecks
}

func NewBackend(url, health string) *Backend {
	return &Backend{
		URL:    url,
		Health: health,
	}
}
