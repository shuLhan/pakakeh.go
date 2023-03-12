// Copyright 2023, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package telemetry

import "context"

type agentContext string

// List of context for agent.
const (
	agentContextForwardWait agentContext = `agent_push_wait`
)

// ContextForwardWait wait for the [Agent.Forward] or [Agent.BulkForward] to be
// finished.
func ContextForwardWait(ctx context.Context) context.Context {
	return context.WithValue(ctx, agentContextForwardWait, struct{}{})
}
