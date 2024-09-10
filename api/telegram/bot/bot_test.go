// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package bot

import (
	"log"
	"os"
	"testing"
)

const (
	testListenAddress = ":1928"
)

var (
	testBot *Bot
)

func TestMain(m *testing.M) {
	startTestBot()

	os.Exit(m.Run())
}

func startTestBot() {
	var err error

	opts := Options{
		HandleUpdate: testHandleUpdate,
		Webhook: &Webhook{
			ListenAddress: testListenAddress,
		},
	}

	testBot, err = New(opts)
	if err != nil {
		log.Println("startTestBot: ", err)
	}

	if testBot != nil {
		go func() {
			err := testBot.Start()
			if err != nil {
				log.Println(err)
			}
		}()
	}
}

func testHandleUpdate(update Update) {
	log.Printf("testHandleUpdate: %+v", update)
}

func TestBot_GetMe(t *testing.T) {
	if testBot == nil {
		t.Skip()
	}

	user, err := testBot.GetMe()
	if err != nil {
		log.Fatal(err)
	}

	t.Logf("GetMe: %+v", user)
}

func TestBot_GetWebhookInfo(t *testing.T) {
	if testBot == nil {
		t.Skip()
	}

	whInfo, err := testBot.GetWebhookInfo()
	if err != nil {
		log.Fatal(err)
	}

	t.Logf("GetWebhookInfo: %+v", whInfo)
}
