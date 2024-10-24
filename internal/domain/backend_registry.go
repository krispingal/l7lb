package domain

type BackendRegistry interface {
	UpdateHealth(status BackendStatus) error
	Subscribe(backendId uint64) <-chan BackendStatus
	GetBackendById(backendId uint64) (Backend, bool)
	AddBackendToRegistry(backend Backend)
}
