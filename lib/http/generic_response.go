package http

import liberrors "github.com/shuLhan/share/lib/errors"

//
// GenericResponse is one of the common HTTP response container that can be
// used by Server implementor.
// Its embed the lib/errors.E type to work seamlessly with Endpoint.Call
// handler for checking the returned error.
//
// See the example below on how to use it with Endpoint.Call handler.
//
type GenericResponse struct {
	liberrors.E
	Data interface{} `json:"data,omitempty"`
}

//
// Unwrap return the error as instance of *liberror.E.
//
func (gr *GenericResponse) Unwrap() (err error) {
	return &gr.E
}
