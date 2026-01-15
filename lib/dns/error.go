// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2024 Shulhan <ms@kilabit.info>

package dns

import "errors"

// errInvalidMessage define an error when raw DNS message cannot be parsed.
var errInvalidMessage = errors.New(`invalid message`)
