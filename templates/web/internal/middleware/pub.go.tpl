package middleware

import (
	"github.com/gone-io/gone"
)

//go:gone
func NewPlantMiddleware() gone.Goner {
	return &PubMiddleware{}
}

// PubMiddleware 公共中间件
type PubMiddleware struct {
	gone.Flag
	gone.Logger `gone:"gone-logger"`
}

func (m *PubMiddleware) Next(ctx *gone.Context) (interface{}, error) {
	m.Infof("public middleware: %s", ctx.Request.URL)
	//todo
	return nil, nil
}
