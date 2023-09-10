// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bot

import (
	"errors"
	"os"
)

const (
	// EnvToken define the environment variable for setting the Telegram
	// Bot token.
	// The environment variable has higher priority than Options parameter
	// that passed in New() function.
	EnvToken = "TELEGRAM_TOKEN"

	// EnvWebhookURL define the environment variable for setting the
	// Telegram Webhook URL.
	// The environment variable has higher priority than Options parameter
	// that passed in New() function.
	EnvWebhookURL = "TELEGRAM_WEBHOOK_URL"
)

const (
	defListenAddress    = ":80"
	defListenAddressTLS = ":443"
)

// UpdateHandler define the handler when Bot receiving updates.
type UpdateHandler func(update Update)

// Options to create new Bot.
type Options struct {
	// Required.  Your Bot authentication token.
	// This option will be overridden by environment variable
	// TELEGRAM_TOKEN.
	Token string

	// Required.  The function that will be called when Bot receiving
	// Updates.
	HandleUpdate UpdateHandler

	// Optional.  Set this options if the Bot want to receive updates
	// using Webhook.
	Webhook *Webhook
}

// init check for required fields and initialize empty fields with default
// value.
func (opts *Options) init() (err error) {
	// Set the Telegram token and Webhook URL from environment, if its not
	// empty.
	env := os.Getenv(EnvToken)
	if len(env) > 0 {
		opts.Token = env
	}
	env = os.Getenv(EnvWebhookURL)
	if len(env) > 0 {
		if opts.Webhook == nil {
			opts.Webhook = &Webhook{}
		}
		opts.Webhook.URL = env
	}

	if len(opts.Token) == 0 {
		return errors.New("empty Token")
	}
	if opts.HandleUpdate == nil {
		return errors.New("field HandleUpdate must be set to non nil")
	}
	if opts.Webhook == nil {
		return errors.New("empty Webhook URL")
	}
	if len(opts.Webhook.URL) == 0 {
		// Even thought empty URL is allowed by API, which
		// means to clear the previous setWebhook, use the
		// DeleteWebhook instead for consistency.
		return errors.New("empty Webhook URL")
	}
	if len(opts.Webhook.ListenAddress) == 0 {
		if opts.Webhook.ListenCertificate == nil {
			opts.Webhook.ListenAddress = defListenAddress
		} else {
			opts.Webhook.ListenAddress = defListenAddressTLS
		}
	}
	return nil
}
