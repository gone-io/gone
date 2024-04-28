package middleware

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/gin"
	"github.com/gone-io/gone/goner/logrus"
)

//go:gone
func NewPlantMiddleware() gone.Goner {
	return &PubMiddleware{}
}

// PubMiddleware 公共中间件
type PubMiddleware struct {
	gone.Flag
	logrus.Logger `gone:"gone-logger"`
}

func (m *PubMiddleware) Next(ctx *gin.Context) (interface{}, error) {
	m.Infof("public middleware: %s", ctx.Request.URL)
	//todo
	return nil, nil
}
