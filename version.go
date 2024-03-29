// Copyright 2020 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dfs

var (
	version = "dev"
	commit  string

	// Version is used to print dfs version
	Version = func() string {
		if commit != "" {
			return version + "-" + commit
		}
		return version
	}()
)
