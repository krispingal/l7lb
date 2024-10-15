package domain

import (
	"net/http"
)

type LoadBalancer interface {
	RouteRequestToGroup(w http.ResponseWriter, r *http.Request, groupId string)
	ListenForHealthUpdates()
}