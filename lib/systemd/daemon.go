// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2026 M. Shulhan <ms@kilabit.info>

package systemd

import (
	"fmt"
	"math"
	"net"
	"os"
	"strconv"

	"golang.org/x/sys/unix"
)

const envListenPID = `LISTEN_PID`
const envListenFDS = `LISTEN_FDS`

// Listeners return the list of [net.Listener] passed by systemd for socket
// activation.
// For each returned [net.Listener], the main program must check if the
// listener address match with expected listen address.
//
//	listenAddress := `127.0.0.1:8080`
//	listeners, _ := systemd.Listeners(true)
//	for _, l := range listeners {
//		if l.Addr().String() == listenAddress {
//			// The listener match with expected address.
//		}
//	}
//
// Program should ignore the listeners or continue with its default address if
// no address matched.
//
// References,
//   - https://github.com/systemd/systemd/blob/v259/src/libsystemd/sd-daemon/sd-daemon.c
//   - https://0pointer.de/blog/projects/socket-activation.html
//   - https://0pointer.de/blog/projects/socket-activation2.html
func Listeners(unsetEnv bool) (list []net.Listener, err error) {
	logp := `Listeners`

	if unsetEnv {
		defer func() {
			_ = os.Unsetenv(envListenPID)
			_ = os.Unsetenv(envListenFDS)
		}()
	}

	v := os.Getenv(envListenPID)
	if v == `` {
		return nil, nil
	}
	listenPid, err := strconv.Atoi(v)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	pid := os.Getpid()
	if listenPid != pid {
		return nil, fmt.Errorf(`%s: mismatch PID, got %d, want %d`,
			logp, listenPid, pid)
	}

	v = os.Getenv(envListenFDS)
	if v == `` {
		return nil, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return nil, fmt.Errorf(`%s: invalid LISTEN_FDS value %s: %w`, logp, v, err)
	}
	if n < 0 {
		return nil, fmt.Errorf(`%s: invalid LISTEN_FDS value %d`, logp, n)
	}
	const listenFDSStart = 3
	if n > math.MaxInt-listenFDSStart {
		return nil, fmt.Errorf(`%s: invalid LISTEN_FDS value %d`, logp, n)
	}

	for fd := listenFDSStart; fd < listenFDSStart+n; fd++ {
		fdptr := uintptr(fd)
		flags, err := unix.FcntlInt(fdptr, unix.F_GETFD, 0)
		if err != nil {
			return nil, fmt.Errorf(`%s: %w`, logp, err)
		}
		newflags := flags | unix.FD_CLOEXEC
		if flags != newflags {
			_, err = unix.FcntlInt(fdptr, unix.F_SETFD, newflags)
			if err != nil {
				return nil, fmt.Errorf(`%s: %w`, logp, err)
			}
		}

		file := os.NewFile(fdptr, ``)
		fileListener, err := net.FileListener(file)
		if err != nil {
			return nil, fmt.Errorf(`%s: %w`, logp, err)
		}
		list = append(list, fileListener)
	}

	return list, nil
}
