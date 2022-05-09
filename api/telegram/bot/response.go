// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bot

import (
	"encoding/json"
	"fmt"

	"github.com/shuLhan/share/lib/errors"
)

const (
	errMigrateToChatID = "The group has been migrated to a supergroup with the specified ID %d"
	errFloodControl    = "Client exceeding flood control, retry after %d seconds"
)

// response is the internal, generic response from API.
type response struct {
	Ok          bool                `json:"ok"`
	Description string              `json:"description"`
	ErrorCode   int                 `json:"error_code"`
	Parameters  *responseParameters `json:"parameters"`
	Result      interface{}         `json:"result"`
}

// unpack the JSON response.
//
// Any non Ok response will be returned as lib/errors.E with following
// mapping: Description become E.Message, ErrorCode become E.Code.
func (res *response) unpack(in []byte) (err error) {
	err = json.Unmarshal(in, res)
	if err != nil {
		return fmt.Errorf("bot: response.unpack: %w", err)
	}
	if !res.Ok {
		var paramsInfo string
		if res.Parameters != nil {
			if res.Parameters.MigrateToChatID != 0 {
				paramsInfo = fmt.Sprintf(errMigrateToChatID,
					res.Parameters.MigrateToChatID)
			}
			if res.Parameters.RetryAfter != 0 {
				paramsInfo += fmt.Sprintf(errFloodControl,
					res.Parameters.RetryAfter)
			}
		}
		return &errors.E{
			Code:    res.ErrorCode,
			Message: res.Description + "." + paramsInfo,
		}
	}
	return nil
}
