// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slack

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	libhttp "github.com/shuLhan/share/lib/http"
)

//
// WebhookClient for slack.
// Use this for long running program that post message every minutes or
// seconds.
//
type WebhookClient struct {
	*libhttp.Client
	webhookPath string
	user        string
	channel     string
}

//
// NewWebhookClient create new slack client that will write the message using
// webhook URL and optional user and channel.
//
func NewWebhookClient(webhookURL, user, channel string) (wcl *WebhookClient, err error) {
	wurl, err := url.Parse(webhookURL)
	if err != nil {
		return nil, fmt.Errorf("NewWebhookClient: %w", err)
	}

	host := fmt.Sprintf("%s://%s", wurl.Scheme, wurl.Host)
	wcl = &WebhookClient{
		Client:      libhttp.NewClient(host, nil, false),
		webhookPath: wurl.Path,
		user:        user,
		channel:     channel,
	}

	wcl.Client.Timeout = 5 * time.Second

	return wcl, nil
}

//
// Post the Message as is.
//
func (wcl *WebhookClient) Post(msg *Message) (err error) {
	if wcl.Client == nil {
		return nil
	}
	httpRes, resBody, err := wcl.PostJSON(wcl.webhookPath, nil, msg)
	if err != nil {
		return fmt.Errorf("WebhookClient.Post: %w", err)
	}
	if httpRes.StatusCode != http.StatusOK {
		return fmt.Errorf("Post: %s: %s\n", httpRes.Status, resBody)
	}
	return nil
}

//
// Write wrap the raw bytes into Message with the user and channel previously
// defined when creating the client, and post it to slack.
//
func (wcl *WebhookClient) Write(b []byte) (n int, err error) {
	if wcl.Client == nil {
		return 0, nil
	}
	msg := &Message{
		Channel:  wcl.channel,
		Username: wcl.user,
		Text:     string(b),
	}
	err = wcl.Post(msg)
	if err != nil {
		return 0, fmt.Errorf("Write: %w", err)
	}
	return len(b), nil
}

//
// Close the client connection.
//
func (wcl *WebhookClient) Close() (err error) {
	wcl.Client = nil
	return nil
}
