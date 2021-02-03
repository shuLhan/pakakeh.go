// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slack

import (
	"log"
	"os"
)

func ExamplePostWebhook() {
	webhookURL := os.Getenv("SLACK_WEBHOOK_URL")
	if len(webhookURL) == 0 {
		return
	}
	msg := &Message{
		Channel:   "test",
		Username:  "Test",
		IconEmoji: ":ghost:",
		Text:      "Hello, world!",
	}
	err := PostWebhook(webhookURL, msg)
	if err != nil {
		log.Fatal(err)
	}
	//Output:
}
