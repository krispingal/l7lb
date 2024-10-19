package domain

type Backend struct {
	URL    string
	Health string
}

func NewBackend(url, health string) *Backend {
	return &Backend{
		URL:    url,
		Health: health,
	}
}
