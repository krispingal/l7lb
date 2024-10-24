package domain

import "sync"

var (
	idCounter uint64
	idMutex   sync.Mutex
)

type Backend struct {
	Id     uint64
	URL    string
	Health string
}

func NewBackend(url string, health string) *Backend {
	idMutex.Lock()
	idCounter++
	id := idCounter
	idMutex.Unlock()
	return &Backend{
		Id:     id,
		URL:    url,
		Health: health,
	}
}
