// Copyright (c) 2015 Janoš Guljaš <janos@resenje.org>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jsonhttp

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// ErrEmptyRequestBody is returned from UnmarshalRequestBody
// when request body is empty either if Content-Length header
// is 0 or JSON decoder returns EOF.
var ErrEmptyRequestBody = errors.New("empty request body")

// UnmarshalRequestBody unmarshals JSON encoded HTTP request body into
// an arbitrary interface. In case of error, it writes appropriate
// JSON-encoded response to http.ResponseWriter, so the calling handler
// should not write new data if this function returns error.
func UnmarshalRequestBody(w http.ResponseWriter, r *http.Request, v interface{}) error {
	defer r.Body.Close()

	if r.Header.Get("Content-Length") == "0" {
		BadRequest(w, "empty request body")
		return ErrEmptyRequestBody
	}
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		switch e := err.(type) {
		case *json.SyntaxError:
			BadRequest(w, fmt.Sprintf("%v (offset %d)", e, e.Offset))
		case *json.UnmarshalTypeError:
			BadRequest(w, fmt.Sprintf("expected json %s value but got %s (offset %d)", e.Type, e.Value, e.Offset))
		default:
			if err == io.EOF {
				err = ErrEmptyRequestBody
			}
			BadRequest(w, err)
		}
		return err
	}
	return nil
}
