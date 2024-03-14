package http

import "net/http"

// ClientResponse contains the response from HTTP client request.
type ClientResponse struct {
	HTTPResponse *http.Response
	Body         []byte
}
