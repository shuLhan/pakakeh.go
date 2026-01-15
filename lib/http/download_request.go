// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2022 Shulhan <ms@kilabit.info>

package http

import "io"

// DownloadRequest define the parameter for [Client.Download] method.
type DownloadRequest struct {
	// Output define where the downloaded resource from server will be
	// writen.
	// This field is required.
	Output io.Writer

	ClientRequest
}
