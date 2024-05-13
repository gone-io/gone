package middleware

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/gin"
	"net/http"
	"web-mysql/internal/interface/domain"
	"web-mysql/internal/pkg/utils"
)

//go:gone
func NewAuthorizeMiddleware() gone.Goner {
	return &AuthorizeMiddleware{}
}

type AuthorizeMiddleware struct {
	gone.Flag
	gone.Logger `gone:"gone-logger"`
	userKey     string `gone:"config,auth.user-key"`
}

func (m *AuthorizeMiddleware) Next(ctx *gone.Context) (any, error) {
	if userId, err := checkToken(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, map[string]any{
			"code": err.Code(),
			"msg":  err.Msg(),
		})
		ctx.Abort()
	} else {
		m.Debugf("userId:%v", userId)
		ctx.Set(m.userKey, userId)
		ctx.Next()
	}
	return nil, nil
}

func checkToken(ctx *gin.Context) (*domain.User, gone.Error) {
	//todo 修改该函数，实现鉴权的相关逻辑

	if ctx.Query("auth") == "pass" {
		return nil, nil
	}
	return nil, gone.NewParameterError("Unauthorized", utils.Unauthorized)
}
