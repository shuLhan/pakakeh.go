// Copyright 2022, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import "io"

// DownloadRequest define the parameter for Client's Download() method.
type DownloadRequest struct {
	// Output define where the downloaded resource from server will be
	// writen.
	// This field is required.
	Output io.Writer

	ClientRequest
}
