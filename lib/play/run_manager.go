// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package play

import "sync"

type runManager struct {
	sidCommand map[string]*command
	sync.Mutex
}

func (runm *runManager) get(sid string) (cmd *command) {
	runm.Lock()
	cmd = runm.sidCommand[sid]
	runm.Unlock()
	return cmd
}

func (runm *runManager) delete(sid string) {
	runm.Lock()
	delete(runm.sidCommand, sid)
	runm.Unlock()
}

func (runm *runManager) store(sid string, cmd *command) {
	runm.Lock()
	runm.sidCommand[sid] = cmd
	runm.Unlock()
}
