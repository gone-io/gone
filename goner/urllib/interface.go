package urllib

import "github.com/imroc/req/v3"

//go:generate sh -c "mockgen -package=urllib github.com/imroc/req/v3 RoundTripper > req_RoundTripper_mock_test.go"

type Client interface {
	R() *req.Request
	C() *req.Client
}

type Res[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg,omitempty"`
	Data T      `json:"data,omitempty"`
}
