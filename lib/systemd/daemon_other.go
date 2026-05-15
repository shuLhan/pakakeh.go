// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2026 M. Shulhan <ms@kilabit.info>

//go:build !linux

package systemd

import (
	"fmt"
	"net"
)

// Listeners returns nil on non-linux operating system.
func Listeners(unsetEnv bool) (list []net.Listener, err error) {
	fmt.Println(`not linux`)
	return nil, nil
}
