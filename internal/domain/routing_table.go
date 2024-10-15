package domain

type RoutingTable interface {
	AddBackendGroup(groupId string, backends []*Backend) error
	RemoveBackendGroup(groupId string) error
	GetBackendsForGroup(groupId string) ([]*Backend, error)
}
