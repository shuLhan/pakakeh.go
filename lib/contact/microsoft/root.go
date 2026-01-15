// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package microsoft

// Root of response.
type Root struct {
	Context  string    `json:"@odata.context"`
	Contacts []Contact `json:"value"`
}
