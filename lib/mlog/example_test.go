// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mlog_test

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"github.com/shuLhan/share/api/slack"
	"github.com/shuLhan/share/lib/mlog"
)

func ExampleMultiLogger() {
	// The following example import package
	// "github.com/shuLhan/share/api/slack".

	var (
		buf   = bytes.Buffer{}
		wouts = []mlog.NamedWriter{
			mlog.NewNamedWriter(`stdout`, os.Stdout),
			mlog.NewNamedWriter(`buffer`, &buf),
		}
		werrs = []mlog.NamedWriter{
			mlog.NewNamedWriter(`stderr`, os.Stdout),
		}
		multilog        = mlog.NewMultiLogger(``, `mlog:`, wouts, werrs)
		slackWebhookURL = os.Getenv("SLACK_WEBHOOK_URL")
	)

	// Create an error writer to slack.
	if len(slackWebhookURL) > 0 {
		// Get the Slack configuration from environment.
		slackChannel := os.Getenv(`SLACK_WEBHOOK_CHANNEL`)
		slackUsername := os.Getenv(`SLACK_WEBHOOK_USERNAME`)

		// Create Slack's client.
		slackc, err := slack.NewWebhookClient(slackWebhookURL, slackUsername, slackChannel)
		if err != nil {
			log.Fatal(err)
		}

		// Forward all errors to Slack client.
		multilog.RegisterErrorWriter(mlog.NewNamedWriter(`slack`, slackc))
	}

	multilog.Outf(`writing to standard output and buffer`)
	multilog.Errf(`writing to standard error and slack`)
	multilog.Close()

	// Try writing to closed mlog.
	multilog.Outf(`writing to standard output and buffer after close`)

	fmt.Println("Output on buffer:", buf.String())

	// Unordered output:
	// mlog: writing to standard output and buffer
	// mlog: writing to standard error and slack
	// Output on buffer: mlog: writing to standard output and buffer
}
