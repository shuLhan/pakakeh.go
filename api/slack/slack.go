// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package slack provide a simple API for sending message to Slack using only
// standard packages.
package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// PostWebhook send a message using "Incoming Webhook".
func PostWebhook(webhookUrl string, msg *Message) (err error) {
	var (
		logp = `PostWebhook`

		payload []byte
	)

	payload, err = json.Marshal(&msg)
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	var res *http.Response

	res, err = http.DefaultClient.Post(webhookUrl, `application/json`, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	if res.StatusCode != http.StatusOK {
		var resBody []byte
		resBody, err = io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf(`%s: %w`, logp, err)
		}
		return fmt.Errorf(`%s: %s: %s`, logp, res.Status, resBody)
	}

	err = res.Body.Close()
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	return nil
}
