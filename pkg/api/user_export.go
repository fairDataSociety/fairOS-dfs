/*
Copyright © 2020 FairOS Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package api

import (
	"net/http"

	"resenje.org/jsonhttp"
)

// UserExportResponse is the json response sent to export user
type UserExportResponse struct {
	Name    string `json:"userName"`
	Address string `json:"address"`
}

// ExportUserHandler godoc
//
//	@Tags         user
//	@Deprecated
//	@Router       /v1/user/export [post]
func (h *Handler) ExportUserHandler(w http.ResponseWriter, r *http.Request) {
	jsonhttp.BadRequest(w, &response{Message: "user export: deprecated"})
}
