package api

import (
	"net/http"

	"resenje.org/jsonhttp"
)

// UserMigrateHandler is the api handler to migrate local user credential to secondary location in swarm
// it takes only one argument
// - password: the password of the user
func (h *Handler) UserMigrateHandler(w http.ResponseWriter, r *http.Request) {
	jsonhttp.BadRequest(w, &response{Message: "user migrate: deprecated"})
}
