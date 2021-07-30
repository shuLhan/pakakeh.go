// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"errors"
	"testing"

	liberrors "github.com/shuLhan/share/lib/errors"
	"github.com/shuLhan/share/lib/test"
)

type myResponse EndpointResponse

//
// The EndpointResponse when returned as error should be able to converted
// to liberrors.E using errors.As().
//
func TestEndpointResponse_errors_As(t *testing.T) {
	myres := &myResponse{
		E: liberrors.E{
			Code:    400,
			Message: "bad request",
		},
	}

	var err error = myres

	epr := &EndpointRequest{
		Error: err,
	}

	errInternal := &liberrors.E{}
	got := errors.As(epr.Error, &errInternal)
	test.Assert(t, "EndpointRequest: errors.As", true, got)
}
