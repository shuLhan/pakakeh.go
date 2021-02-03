// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//
// Package slack provide a simple API for sending message to Slack using only
// standard packages.
//
package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

//
// PostWebhook send a message using "Incoming Webhook".
//
func PostWebhook(webhookUrl string, msg *Message) (err error) {
	payload, err := json.Marshal(&msg)
	if err != nil {
		return fmt.Errorf("PostWebhook: %w", err)
	}

	res, err := http.DefaultClient.Post(webhookUrl, "application/json",
		bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("PostWebhook: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		resBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("PostWebhook: %w", err)
		}
		return fmt.Errorf("PostWebhook: %s: %s\n", res.Status, resBody)
	}

	err = res.Body.Close()
	if err != nil {
		return fmt.Errorf("PostWebhook: %w", err)
	}

	return nil
}
