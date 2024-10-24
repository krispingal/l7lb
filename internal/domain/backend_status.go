package domain

type BackendStatus struct {
	Id        uint64
	URL       string
	IsHealthy bool
}
