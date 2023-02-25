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

type myResponse struct {
	EndpointResponse
}

func (myres *myResponse) Error() string {
	return myres.EndpointResponse.Error()
}

// The EndpointResponse when returned as error should be able to converted
// to liberrors.E using errors.As().
func TestEndpointResponse_errors_As(t *testing.T) {
	myres := &myResponse{
		EndpointResponse: EndpointResponse{
			E: liberrors.E{
				Code:    400,
				Message: "bad request",
			},
		},
	}

	var err error = myres

	epr := &EndpointRequest{
		Error: err,
	}

	var (
		errIn  *liberrors.E
		errIn2 = &liberrors.E{}
	)

	got := errors.As(epr.Error, &errIn)
	test.Assert(t, "EndpointRequest: errors.As", true, got)
	test.Assert(t, "errors.As with unintialized E:", errIn, &myres.E)

	got = errors.As(epr.Error, &errIn2)
	test.Assert(t, "EndpointRequest: errors.As", true, got)
	test.Assert(t, "errors.As with initialized E:", errIn2, &myres.E)
}
