package domain

type BackendRegistry interface {
	UpdateHealth(status BackendStatus) error
	Subscribe(backendUrl string) <-chan BackendStatus
	GetBackendById(backendId uint64) (Backend, bool)
	AddBackendToRegistry(backend Backend)
}
