// Copyright 2022, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package test

import "testing"

func TestData_parse(t *testing.T) {
	type testCase struct {
		desc    string
		content []byte
		expData Data
	}

	var cases = []testCase{{
		desc:    `With flag only`,
		content: []byte("\na: b\nc: d\n"),
		expData: Data{
			Flag: map[string]string{
				`a`: `b`,
				`c`: `d`,
			},
		},
	}, {
		desc:    `With description only`,
		content: []byte("\nDesc."),
		expData: Data{
			Desc: []byte("Desc."),
		},
	}, {
		desc:    `With input only`,
		content: []byte(">>>\n\ninput.\n\n"),
		expData: Data{
			Input: map[string][]byte{
				`default`: []byte("\ninput.\n\n"),
			},
		},
	}, {
		desc:    `With output only`,
		content: []byte("<<<\n\noutput.\n\n"),
		expData: Data{
			Output: map[string][]byte{
				`default`: []byte("\noutput.\n\n"),
			},
		},
	}, {
		desc:    `With flag and description`,
		content: []byte("a: b\nMulti\nline\ndescription.\n"),
		expData: Data{
			Flag: map[string]string{
				`a`: `b`,
			},
			Desc: []byte("Multi\nline\ndescription."),
		},
	}, {
		desc: `With multi input`,
		content: []byte("a: b\n" +
			"Desc.\n" +
			">>> input 1\n1\n\n" +
			">>> input 2\n2\n",
		),
		expData: Data{
			Flag: map[string]string{
				`a`: `b`,
			},
			Desc: []byte("Desc."),
			Input: map[string][]byte{
				"input 1": []byte("1\n"),
				"input 2": []byte("2\n"),
			},
		},
	}, {
		desc: `With multi output`,
		content: []byte("Desc.\n" +
			"<<< output-1\n1\n\n2\n\n" +
			"<<< output-2\n3\n\n4\n",
		),
		expData: Data{
			Flag: map[string]string{},
			Desc: []byte("Desc."),
			Output: map[string][]byte{
				"output-1": []byte("1\n\n2\n"),
				"output-2": []byte("3\n\n4\n"),
			},
		},
	}, {
		desc: `With input duplicate names`,
		content: []byte(">>>\n" +
			"input 1\n\n" +
			">>> default\nInput 2.\n",
		),
		expData: Data{
			Flag: map[string]string{},
			Input: map[string][]byte{
				"default": []byte("Input 2.\n"),
			},
		},
	}, {
		desc: `With no newline above output`,
		content: []byte(">>>\n" +
			"Input 1.\n" +
			"<<<\n" +
			"Output 1.\n",
		),
		expData: Data{
			Input: map[string][]byte{
				"default": []byte("Input 1.\n<<<\nOutput 1.\n"),
			},
		},
	}}

	var (
		c       testCase
		gotData *Data
		err     error
	)

	for _, c = range cases {
		t.Run(c.desc, func(t *testing.T) {
			gotData = newData("")
			err = gotData.parse(c.content)
			if err != nil {
				t.Fatal(err)
			}

			Assert(t, "Data", &c.expData, gotData)
		})
	}
}
