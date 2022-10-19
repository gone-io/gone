package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/logrus"
	"github.com/gone-io/gone/goner/tracer"
	"net/http"
)

// NewGinResponser 新建系统默认的响应处理器
// 注入的ID为：gone-gin-responser (`gone.IdGoneGinResponser`)
func NewGinResponser() (gone.Goner, gone.GonerId) {
	return &responser{}, gone.IdGoneGinResponser
}

type res map[string]interface{}

type responser struct {
	gone.Flag
	logrus.Logger `gone:"gone-logger"`
	tracer        tracer.Tracer `gone:"gone-tracer"`
}

func (r *responser) Success(ctx *gin.Context, data interface{}) {
	bErr, ok := data.(BusinessError)
	if ok {
		ctx.JSON(http.StatusOK, res{
			"code": bErr.Code(),
			"msg":  bErr.Msg(),
			"data": bErr.Data(),
		})
		return
	}
	ctx.JSON(http.StatusOK, res{
		"code": 0,
		"data": data,
	})
}

func (r *responser) Failed(ctx *gin.Context, oErr error) {
	err := ToError(oErr)
	businessError, ok := err.(BusinessError)
	if ok {
		ctx.JSON(http.StatusOK, res{
			"code": businessError.Code(),
			"msg":  businessError.Msg(),
			"data": businessError.Data(),
		})
		return
	}

	iErr, ok := err.(*iError)
	if ok {
		ctx.JSON(http.StatusInternalServerError, res{
			"code": iErr.Code(),
			"msg":  iErr.Msg(),
		})
		r.tracer.Go(func() {
			if ok {
				r.Errorf("inner Error: %s(code=%d)\n%s", iErr.msg, iErr.code, iErr.stack)
			}
		})
		return
	}
	ctx.JSON(http.StatusBadRequest, res{
		"code": err.Code(),
		"msg":  err.Msg(),
	})
}
