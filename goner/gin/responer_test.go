package gin_test

import (
	"encoding/json"
	"errors"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner"
	"github.com/gone-io/gone/goner/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

type jsonSender struct {
	Code int
	Obj  map[string]interface{}
}

func (j *jsonSender) JSON(code int, obj any) {
	j.Code = code
	js, _ := json.Marshal(obj)
	_ = json.Unmarshal(js, &j.Obj)
}

func Test_responserFailed(t *testing.T) {
	gone.Test(func(responser gin.Responser) {
		ctx := jsonSender{}

		responser.Success(&ctx, map[string]interface{}{"ok": 1})
		assert.Equal(t, ctx.Code, http.StatusOK)
		m, ok := ctx.Obj["data"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, m["ok"], float64(1))

		responser.Failed(&ctx, errors.New("my test error"))
		assert.Equal(t, ctx.Code, http.StatusInternalServerError)
		assert.Equal(t, ctx.Obj["code"], float64(http.StatusInternalServerError))

		responser.Failed(&ctx, gin.NewParameterError("test", 100))
		assert.Equal(t, ctx.Code, http.StatusBadRequest)
		assert.Equal(t, ctx.Obj["msg"], "test")
		assert.Equal(t, ctx.Obj["code"], float64(100))

		responser.Failed(&ctx, gin.NewInnerError("test", 100))
		assert.Equal(t, ctx.Code, http.StatusInternalServerError)
		assert.Equal(t, ctx.Obj["code"], float64(100))

		responser.Failed(&ctx, gin.NewBusinessError("depends", 200, map[string]interface{}{"depends": 10}))
		assert.Equal(t, ctx.Code, http.StatusOK)
		assert.Equal(t, ctx.Obj["msg"], "depends")
		assert.Equal(t, ctx.Obj["code"], float64(200))
		m, ok = ctx.Obj["data"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, m["depends"], float64(10))

	}, goner.BasePriest, func(cemetery gone.Cemetery) error {
		cemetery.Bury(gin.NewGinResponser())
		return nil
	})
}
