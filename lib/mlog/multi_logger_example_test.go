// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mlog

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"github.com/shuLhan/share/api/slack"
)

func ExampleMultiLogger() {
	buf := bytes.Buffer{}

	wouts := []NamedWriter{
		NewNamedWriter("stdout", os.Stdout),
		NewNamedWriter("buffer", &buf),
	}
	werrs := []NamedWriter{
		NewNamedWriter("stderr", os.Stdout),
	}

	mlog := NewMultiLogger("", "mlog:", wouts, werrs)

	// Create an error writer to slack.
	slackWebhookURL := os.Getenv("SLACK_WEBHOOK_URL")
	if len(slackWebhookURL) > 0 {
		slackChannel := os.Getenv("SLACK_WEBHOOK_CHANNEL")
		slackUsername := os.Getenv("SLACK_WEBHOOK_USERNAME")

		slackc, err := slack.NewWebhookClient(slackWebhookURL, slackUsername, slackChannel)
		if err != nil {
			log.Fatal(err)
		}
		mlog.RegisterErrorWriter(NewNamedWriter("slack", slackc))
	}

	mlog.Outf("writing to standard output and buffer\n")
	mlog.Errf("writing to standard error and slack\n")
	mlog.Flush()
	fmt.Printf("Output on buffer: %s\n", buf.String())

	// Unordered output:
	// mlog: writing to standard output and buffer
	// mlog: writing to standard error and slack
	// Output on buffer: mlog: writing to standard output and buffer
}
