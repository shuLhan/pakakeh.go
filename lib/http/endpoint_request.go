package http

import "net/http"

//
// EndpointRequest wrap the called Endpoint and common two parameters in HTTP
// handler: the http.ResponseWriter and http.Request.
//
// The RequestBody field contains the full http.Request.Body that has been
// read.
//
// The Error field is used by CallbackErrorHandler.
//
type EndpointRequest struct {
	Endpoint    *Endpoint
	HttpWriter  http.ResponseWriter
	HttpRequest *http.Request
	RequestBody []byte
	Error       error
}
