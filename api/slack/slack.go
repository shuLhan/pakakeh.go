// SPDX-FileCopyrightText: 2021 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

// Package slack provide a simple API for sending message to Slack using only
// standard packages.
package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// PostWebhook send a message using "Incoming Webhook".
func PostWebhook(webhookURL string, msg *Message) (err error) {
	var (
		logp = `PostWebhook`

		payload []byte
	)

	payload, err = json.Marshal(&msg)
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	var (
		ctx     = context.Background()
		httpReq *http.Request
	)

	httpReq, err = http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	httpReq.Header.Set(`Content-Type`, `application/json`)

	var res *http.Response

	res, err = http.DefaultClient.Do(httpReq)
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
