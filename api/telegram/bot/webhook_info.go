// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bot

// WebhookInfo contains information about the current status of a webhook.
type WebhookInfo struct {
	// Webhook URL, may be empty if webhook is not set up.
	URL string `json:"url"`

	// Optional. Error message in human-readable format for the most
	// recent error that happened when trying to deliver an update via
	// webhook.
	LastErrorMessage string `json:"last_error_message"`

	// Optional. A list of update types the bot is subscribed to. Defaults
	// to all update types.
	AllowedUpdates []string `json:"allowed_updates"`

	// Number of updates awaiting delivery
	PendingUpdateCount int `json:"pending_update_count"`

	// Optional. Unix time for the most recent error that happened when
	// trying to deliver an update via webhook.
	LastErrorDate int `json:"last_error_date"`

	// Optional. Maximum allowed number of simultaneous HTTPS connections
	// to the webhook for update delivery.
	MaxConnections int `json:"max_connections"`

	// True, if a custom certificate was provided for webhook certificate
	// checks.
	HasCustomCertificate bool `json:"has_custom_certificate"`
}
