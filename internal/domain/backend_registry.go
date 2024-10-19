package domain

type BackendRegistry interface {
	UpdateHealth(status BackendStatus) error
	Subscribe(backendUrl string) <-chan BackendStatus
	GetBackendByURL(backendUrl string) (Backend, bool)
	AddBackendToRegistry(backend Backend)
}
