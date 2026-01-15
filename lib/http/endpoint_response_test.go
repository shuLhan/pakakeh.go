// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2021 Shulhan <ms@kilabit.info>

package http

import (
	"errors"
	"testing"

	liberrors "git.sr.ht/~shulhan/pakakeh.go/lib/errors"
	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
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
