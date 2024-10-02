package domain

type Backend struct {
	URL    string
	Alive  bool
	Health string
}

func NewBackend(url, health string) *Backend {
	return &Backend{
		URL:    url,
		Alive:  true,
		Health: health,
	}
}
