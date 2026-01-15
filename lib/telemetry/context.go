// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2023 Shulhan <ms@kilabit.info>

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
