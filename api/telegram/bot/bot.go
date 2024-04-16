// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bot

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	"git.sr.ht/~shulhan/pakakeh.go/lib/errors"
	libhttp "git.sr.ht/~shulhan/pakakeh.go/lib/http"
)

// List of message parse mode.
const (
	ParseModeMarkdownV2 = "MarkdownV2"
	ParseModeHTML       = "HTML"
)

// List of Update types.
//
// This types can be used to set AllowedUpdates on Options.Webhook.
const (
	// New incoming message of any kind — text, photo, sticker, etc.
	UpdateTypeMessage = "message"

	// New version of a message that is known to the bot and was edited.
	UpdateTypeEditedMessage = "edited_message"

	// New incoming channel post of any kind — text, photo, sticker, etc.
	UpdateTypeChannelPost = "channel_post"

	// New version of a channel post that is known to the bot and was
	// edited.
	UpdateTypeEditedChannelPost = "edited_channel_post"

	// New incoming inline query
	UpdateTypeInlineQuery = "inline_query"

	// The result of an inline query that was chosen by a user and sent to
	// their chat partner.
	UpdateTypeChosenInlineResult = "chosen_inline_result"

	// New incoming callback query.
	UpdateTypeCallbackQuery = "callback_query"

	// New incoming shipping query.
	// Only for invoices with flexible price.
	UpdateTypeShippingQuery = "shipping_query"

	// New incoming pre-checkout query.
	// Contains full information about checkout.
	UpdateTypePreCheckoutQuery = "pre_checkout_query"

	// New poll state.
	// Bots receive only updates about stopped polls and polls, which are
	// sent by the bot.
	UpdateTypePoll = "poll"

	// A user changed their answer in a non-anonymous poll.
	// Bots receive new votes only in polls that were sent by the bot
	// itself.
	UpdateTypePollAnswer = "poll_answer"
)

const (
	defURL = "https://api.telegram.org/bot"
)

// List of API methods.
const (
	methodDeleteWebhook  = "deleteWebhook"
	methodGetMe          = "getMe"
	methodGetMyCommands  = "getMyCommands"
	methodGetWebhookInfo = "getWebhookInfo"
	methodSendMessage    = "sendMessage"
	methodSetMyCommands  = "setMyCommands"
	methodSetWebhook     = "setWebhook"
)

const (
	paramNameURL            = "url"
	paramNameCertificate    = "certificate"
	paramNameMaxConnections = "max_connections"
	paramNameAllowedUpdates = "allowed_updates"
)

// Bot for Telegram using webHook.
type Bot struct {
	opts     Options
	client   *libhttp.Client
	webhook  *libhttp.Server
	user     *User
	err      chan error
	commands commands
}

// New create and initialize new Telegram bot.
func New(opts Options) (bot *Bot, err error) {
	err = opts.init()
	if err != nil {
		return nil, fmt.Errorf("bot.New: %w", err)
	}

	var clientOpts = libhttp.ClientOptions{
		ServerURL: defURL + opts.Token + `/`,
	}
	bot = &Bot{
		opts:   opts,
		client: libhttp.NewClient(clientOpts),
	}

	fmt.Printf("Bot options: %+v\n", opts)
	fmt.Printf("Bot options Webhook: %+v\n", opts.Webhook)

	// Check if Bot Token is correct by issuing "getMe" method to API
	// server.
	bot.user, err = bot.GetMe()
	if err != nil {
		return nil, err
	}

	return bot, nil
}

// DeleteWebhook remove webhook integration if you decide to switch back to
// getUpdates. Returns True on success. Requires no parameters.
func (bot *Bot) DeleteWebhook() (err error) {
	var (
		req = libhttp.ClientRequest{
			Path: methodDeleteWebhook,
		}
		clientRes *libhttp.ClientResponse
	)

	clientRes, err = bot.client.PostForm(req)
	if err != nil {
		return fmt.Errorf("DeleteWebhook: %w", err)
	}

	var res = &response{}
	err = json.Unmarshal(clientRes.Body, res)
	if err != nil {
		return fmt.Errorf("DeleteWebhook: %w", err)
	}

	return nil
}

// GetMe A simple method for testing your bot's auth token.
// Requires no parameters.
// Returns basic information about the bot in form of a User object.
func (bot *Bot) GetMe() (user *User, err error) {
	var (
		req = libhttp.ClientRequest{
			Path: methodGetMe,
		}
		clientRes *libhttp.ClientResponse
	)

	clientRes, err = bot.client.Get(req)
	if err != nil {
		return nil, fmt.Errorf("GetMe: %w", err)
	}

	user = &User{}
	var res = &response{
		Result: user,
	}
	err = res.unpack(clientRes.Body)
	if err != nil {
		return nil, fmt.Errorf("GetMe: %w", err)
	}

	return user, nil
}

// GetMyCommands get the current list of the bot's commands.
func (bot *Bot) GetMyCommands() (cmds []Command, err error) {
	var (
		req = libhttp.ClientRequest{
			Path: methodGetMyCommands,
		}
		clientRes *libhttp.ClientResponse
	)

	clientRes, err = bot.client.Get(req)
	if err != nil {
		return nil, fmt.Errorf("GetMyCommands: %w", err)
	}

	res := &response{
		Result: cmds,
	}
	err = res.unpack(clientRes.Body)
	if err != nil {
		return nil, fmt.Errorf("GetMyCommands: %w", err)
	}

	return cmds, nil
}

// GetWebhookInfo get current webhook status. Requires no parameters.
// On success, returns a WebhookInfo object.
// If the bot is using getUpdates, will return an object with the url field
// empty.
func (bot *Bot) GetWebhookInfo() (webhookInfo *WebhookInfo, err error) {
	var (
		req = libhttp.ClientRequest{
			Path: methodGetWebhookInfo,
		}
		clientRes *libhttp.ClientResponse
	)

	clientRes, err = bot.client.Get(req)
	if err != nil {
		return nil, fmt.Errorf("GetWebhookInfo: %w", err)
	}

	webhookInfo = &WebhookInfo{}
	res := &response{
		Result: webhookInfo,
	}
	err = res.unpack(clientRes.Body)
	if err != nil {
		return nil, fmt.Errorf("GetWebhookInfo: %w", err)
	}

	return webhookInfo, nil
}

// SendMessage send text messages using defined parse mode to specific
// user.
func (bot *Bot) SendMessage(parent *Message, parseMode, text string) (
	msg *Message, err error,
) {
	var (
		params = messageRequest{
			ChatID:    parent.Chat.ID,
			Text:      text,
			ParseMode: parseMode,
		}
		req = libhttp.ClientRequest{
			Path:   methodSendMessage,
			Params: params,
		}
		clientRes *libhttp.ClientResponse
	)

	clientRes, err = bot.client.PostJSON(req)
	if err != nil {
		return nil, fmt.Errorf("SendMessage: %w", err)
	}

	msg = &Message{}
	res := response{
		Result: msg,
	}
	err = res.unpack(clientRes.Body)
	if err != nil {
		return nil, fmt.Errorf("SendMessage: %w", err)
	}

	return msg, nil
}

// SetMyCommands change the list of the bot's commands.
//
// The value of each Command in the list must be valid according to
// description in Command type; this is including length and characters.
func (bot *Bot) SetMyCommands(cmds []Command) (err error) {
	if len(cmds) == 0 {
		return nil
	}
	for _, cmd := range cmds {
		err = cmd.validate()
		if err != nil {
			return fmt.Errorf("SetMyCommands: %w", err)
		}
	}

	bot.commands.Commands = cmds

	var (
		req = libhttp.ClientRequest{
			Path:   methodSetMyCommands,
			Params: &bot.commands,
		}
		clientRes *libhttp.ClientResponse
	)

	clientRes, err = bot.client.PostJSON(req)
	if err != nil {
		return fmt.Errorf("SetMyCommands: %w", err)
	}

	res := &response{}
	err = res.unpack(clientRes.Body)
	if err != nil {
		return fmt.Errorf("SetMyCommands: %w", err)
	}

	return nil
}

// Start the Bot.
//
// If the Webhook option is not nil it will start listening to updates through
// webhook.
func (bot *Bot) Start() (err error) {
	if bot.opts.Webhook != nil {
		return bot.startWebhook()
	}
	return nil
}

// Stop the Bot.
func (bot *Bot) Stop() (err error) {
	if bot.webhook != nil {
		err = bot.webhook.Shutdown(context.TODO())
		if err != nil {
			log.Println("bot: Stop: ", err)
		}

		bot.webhook = nil
	}

	return nil
}

func (bot *Bot) setWebhook() (err error) {
	var (
		logp       = `setWebhook`
		params     = make(map[string][]byte)
		webhookURL = bot.opts.Webhook.URL + path.Join(`/`, bot.opts.Token)
	)

	params[paramNameURL] = []byte(webhookURL)
	if len(bot.opts.Webhook.Certificate) > 0 {
		params[paramNameCertificate] = bot.opts.Webhook.Certificate
	}
	if bot.opts.Webhook.MaxConnections > 0 {
		var str = strconv.Itoa(bot.opts.Webhook.MaxConnections)
		params[paramNameMaxConnections] = []byte(str)
	}
	if len(bot.opts.Webhook.AllowedUpdates) > 0 {
		var allowedUpdates []byte
		allowedUpdates, err = json.Marshal(&bot.opts.Webhook.AllowedUpdates)
		if err != nil {
			return fmt.Errorf(`%s: %w`, logp, err)
		}
		params[paramNameAllowedUpdates] = allowedUpdates
	}

	var (
		req = libhttp.ClientRequest{
			Path:   methodSetWebhook,
			Params: params,
		}
		clientRes *libhttp.ClientResponse
	)

	clientRes, err = bot.client.PostFormData(req)
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	var res = &response{}

	err = json.Unmarshal(clientRes.Body, res)
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	fmt.Printf("%s: response: %+v\n", logp, res)

	return nil
}

// startWebhook start the HTTP server to receive Update from Telegram API
// server and register the Webhook.
func (bot *Bot) startWebhook() (err error) {
	err = bot.createServer()
	if err != nil {
		return fmt.Errorf("startWebhook: %w", err)
	}

	bot.err = make(chan error)

	go func() {
		bot.err <- bot.webhook.Start()
	}()

	err = bot.DeleteWebhook()
	if err != nil {
		log.Println("startWebhook:", err.Error())
	}

	err = bot.setWebhook()
	if err != nil {
		return fmt.Errorf("startWebhook: %w", err)
	}

	return <-bot.err
}

// createServer start the HTTP server for receiving Update.
func (bot *Bot) createServer() (err error) {
	var logp = `createServer`

	var serverOpts = libhttp.ServerOptions{
		Address: bot.opts.Webhook.ListenAddress,
	}

	if bot.opts.Webhook.ListenCertificate != nil {
		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
		tlsConfig.Certificates = append(
			tlsConfig.Certificates,
			*bot.opts.Webhook.ListenCertificate,
		)
		serverOpts.Conn = &http.Server{
			TLSConfig:         tlsConfig,
			ReadHeaderTimeout: 5 * time.Second,
		}
	}

	bot.webhook, err = libhttp.NewServer(serverOpts)
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	var webhookURL *url.URL

	webhookURL, err = url.Parse(bot.opts.Webhook.URL)
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	var fullPath = path.Join(webhookURL.Path, bot.opts.Token)

	var ep = libhttp.Endpoint{
		Method:       libhttp.RequestMethodGet,
		Path:         fullPath,
		RequestType:  libhttp.RequestTypeNone,
		ResponseType: libhttp.ResponseTypeJSON,
		Call:         bot.handleWebhookGet,
	}

	err = bot.webhook.RegisterEndpoint(ep)
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	ep = libhttp.Endpoint{
		Method:       libhttp.RequestMethodPost,
		Path:         fullPath,
		RequestType:  libhttp.RequestTypeJSON,
		ResponseType: libhttp.ResponseTypeNone,
		Call:         bot.handleWebhook,
	}

	err = bot.webhook.RegisterEndpoint(ep)
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	return nil
}

// handleWebhook handle Updates from Webhook.
func (bot *Bot) handleWebhook(epr *libhttp.EndpointRequest) (resBody []byte, err error) {
	var update Update

	err = json.Unmarshal(epr.RequestBody, &update)
	if err != nil {
		return nil, errors.Internal(err)
	}

	var isHandled bool

	if len(bot.commands.Commands) > 0 && update.Message != nil {
		isHandled = bot.handleUpdateCommand(update)
	}

	// If no Command handler found, forward it to global handler.
	if !isHandled {
		bot.opts.HandleUpdate(update)
	}

	return resBody, nil
}

func (bot *Bot) handleWebhookGet(_ *libhttp.EndpointRequest) (resBody []byte, err error) {
	var res = libhttp.EndpointResponse{}
	res.Code = 200
	res.Message = `OK`
	resBody, err = json.Marshal(&res)
	if err != nil {
		return nil, err
	}
	return resBody, nil
}

func (bot *Bot) handleUpdateCommand(update Update) bool {
	ok := update.Message.parseCommandArgs()
	if !ok {
		return false
	}

	for _, cmd := range bot.commands.Commands {
		if cmd.Command == update.Message.Command {
			if cmd.Handler != nil {
				cmd.Handler(update)
			}
			return true
		}
	}
	return false
}
