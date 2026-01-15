// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2022 Shulhan <ms@kilabit.info>

package dns

import "fmt"

func ExampleHostsFile_Get() {
	var (
		hostsFile = &HostsFile{
			Records: []*ResourceRecord{{
				Name:  "my.local",
				Type:  RecordTypeA,
				Value: "127.0.0.1",
			}, {
				Name:  "my.local",
				Type:  RecordTypeA,
				Value: "127.0.0.2",
			}},
		}
	)

	fmt.Println(hostsFile.Get("my.local", ""))
	fmt.Println(hostsFile.Get("my.local", "127.0.0.2"))
	fmt.Println(hostsFile.Get("my.my", ""))
	// Output:
	// {Name:my.local Type:1 Class:0 TTL:0 Value:127.0.0.1}
	// {Name:my.local Type:1 Class:0 TTL:0 Value:127.0.0.2}
	// <nil>
}
