package http

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func ExampleEndpoint_errorHandler() {
	serverOpts := &ServerOptions{
		Address: "127.0.0.1:8123",
	}
	server, _ := NewServer(serverOpts)

	endpointError := &Endpoint{
		Method:       RequestMethodGet,
		Path:         "/",
		RequestType:  RequestTypeQuery,
		ResponseType: ResponseTypePlain,
		Call: func(res http.ResponseWriter, req *http.Request, reqBody []byte) ([]byte, error) {
			return nil, fmt.Errorf(req.Form.Get("error"))
		},
		ErrorHandler: func(res http.ResponseWriter, req *http.Request, err error) {
			res.Header().Set(HeaderContentType, ContentTypePlain)

			codeMsg := strings.Split(err.Error(), ":")
			if len(codeMsg) != 2 {
				res.WriteHeader(http.StatusInternalServerError)
				res.Write([]byte(err.Error()))
			} else {
				code, _ := strconv.Atoi(codeMsg[0])
				res.WriteHeader(code)
				res.Write([]byte(codeMsg[1]))
			}
		},
	}
	_ = server.RegisterEndpoint(endpointError)

	go func() {
		_ = server.Start()
	}()
	defer server.Stop(1 * time.Second)
	time.Sleep(1 * time.Second)

	client := NewClient("http://"+serverOpts.Address, nil, false)

	params := url.Values{}
	params.Set("error", "400:error with status code")
	httpres, resbody, _ := client.Get(nil, "/", params)
	fmt.Printf("%d: %s\n", httpres.StatusCode, resbody)

	params.Set("error", "error without status code")
	httpres, resbody, _ = client.Get(nil, "/", params)
	fmt.Printf("%d: %s\n", httpres.StatusCode, resbody)

	// Output:
	// 400: error with status code
	// 500: error without status code
}
