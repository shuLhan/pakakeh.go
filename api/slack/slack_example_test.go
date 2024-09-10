// SPDX-FileCopyrightText: 2021 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

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
