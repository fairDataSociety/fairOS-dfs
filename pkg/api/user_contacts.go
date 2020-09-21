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

	"github.com/fairdatasociety/fairOS-dfs/pkg/cookie"
	"github.com/fairdatasociety/fairOS-dfs/pkg/user"
)

func (h *Handler) SaveUserContactHandler(w http.ResponseWriter, r *http.Request) {
	phone := r.FormValue("phone")
	mobile := r.FormValue("mobile")

	addrLine1 := r.FormValue("address_line_1")
	addrLine2 := r.FormValue("address_line_2")
	if addrLine1 != "" && addrLine2 == "" {
		h.logger.Errorf("user save contact: \"address_line_2\" argument missing")
		jsonhttp.BadRequest(w, "user save contact: \"address_line_2\" argument missing")
		return
	}
	state := r.FormValue("state_province_region")
	if addrLine1 != "" && state == "" {
		h.logger.Errorf("user save contact: \"state_province_region\" argument missing")
		jsonhttp.BadRequest(w, "user save contact: \"state_province_region\" argument missing")
		return
	}
	zipCode := r.FormValue("zipcode")
	if addrLine1 != "" && zipCode == "" {
		h.logger.Errorf("user save contact: \"zipcode\" argument missing")
		jsonhttp.BadRequest(w, "user save contact: \"zipcode\" argument missing")
		return
	}

	if phone == "" && mobile == "" && addrLine1 == "" {
		h.logger.Errorf("user save contact: one of the contact information should be given")
		jsonhttp.BadRequest(w, "user save contact: one of the contact information should be given")
		return
	}

	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("user save contact: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("user save contact: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "user save contact: \"cookie-id\" parameter missing in cookie")
		return
	}

	var address *user.Address
	if addrLine1 != "" {
		address = &user.Address{
			AddressLine1: addrLine1,
			AddressLine2: addrLine2,
			State:        state,
			ZipCode:      zipCode,
		}
	}

	err = h.dfsAPI.SaveContact(phone, mobile, address, sessionId)
	if err != nil {
		h.logger.Errorf("user save contact: %v", err)
		jsonhttp.InternalServerError(w, "user save contact: "+err.Error())
		return
	}
	jsonhttp.OK(w, nil)
}

func (h *Handler) GetUserContactHandler(w http.ResponseWriter, r *http.Request) {
	// get values from cookie
	sessionId, err := cookie.GetSessionIdFromCookie(r)
	if err != nil {
		h.logger.Errorf("user get contact: invalid cookie: %v", err)
		jsonhttp.BadRequest(w, ErrInvalidCookie)
		return
	}
	if sessionId == "" {
		h.logger.Errorf("user get contact: \"cookie-id\" parameter missing in cookie")
		jsonhttp.BadRequest(w, "user get contact: \"cookie-id\" parameter missing in cookie")
		return
	}

	contacts, err := h.dfsAPI.GetContact(sessionId)
	if err != nil {
		h.logger.Errorf("user get contact: %v", err)
		jsonhttp.InternalServerError(w, "user get contact: "+err.Error())
		return
	}
	jsonhttp.OK(w, contacts)
}
