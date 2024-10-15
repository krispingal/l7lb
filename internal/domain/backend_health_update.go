package domain

type BackendHealthUpdate struct {
	Backend   *Backend
	IsHealthy bool
	GroupId string
}
