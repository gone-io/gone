package gin

import (
	"encoding/json"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"strings"
	"testing"
)

func TestNewGinResponser(t *testing.T) {
	_, _, _ = NewGinResponser()
}

func (r *responser) Go(func())                         {}
func (r *responser) Warnf(format string, args ...any)  {}
func (r *responser) Errorf(format string, args ...any) {}

func Test_responser_Success(t *testing.T) {
	t.Run("returnWrappedData=false", func(t *testing.T) {
		controller := gomock.NewController(t)
		defer controller.Finish()

		r := responser{
			wrappedDataFunc: wrapFunc,
		}

		ctx := NewMockXContext(controller)
		t.Run("data is struct|map|slice|array", func(t *testing.T) {
			var x struct {
				X string
			}
			x.X = "my-test"

			ctx.EXPECT().JSON(gomock.Any(), gomock.Any()).Do(func(code int, obj any) {
				assert.Equal(t, x, obj)
			})
			r.Success(ctx, x)
		})

		t.Run("data is pointer to struct|map|slice|array", func(t *testing.T) {
			var x struct {
				X string
			}
			x.X = "my-test"

			ctx.EXPECT().JSON(gomock.Any(), gomock.Any()).Do(func(code int, obj any) {
				assert.Equal(t, &x, obj)
			})
			r.Success(ctx, &x)
		})

		t.Run("data is pointer to other", func(t *testing.T) {
			x := "my-test"

			ctx.EXPECT().String(gomock.Any(), gomock.Any()).Do(func(code int, format string, values ...any) {
				assert.Equal(t, x, format)
			})
			r.Success(ctx, &x)
		})

		t.Run("data is string", func(t *testing.T) {
			ctx.EXPECT().String(gomock.Any(), gomock.Any()).Do(func(code int, format string, values ...any) {
				assert.Equal(t, "test", format)
			})
			r.Success(ctx, "test")
		})

		t.Run("data is nil", func(t *testing.T) {
			ctx.EXPECT().String(gomock.Any(), gomock.Any()).Do(func(code int, format string, values ...any) {
				assert.Equal(t, "", format)
			})
			r.Success(ctx, nil)
		})
	})

	t.Run("returnWrappedData=false", func(t *testing.T) {
		controller := gomock.NewController(t)
		defer controller.Finish()

		r := responser{
			wrappedDataFunc:   wrapFunc,
			returnWrappedData: true,
		}

		ctx := NewMockXContext(controller)

		t.Run("data is BusinessError", func(t *testing.T) {
			x := NewBusinessError("test", 1, "test-data")

			ctx.EXPECT().JSON(gomock.Any(), gomock.Any()).Do(func(code int, obj any) {
				marshal, err := json.Marshal(obj)
				assert.Nil(t, err)

				assert.Equal(t, "{\"code\":1,\"msg\":\"test\",\"data\":\"test-data\"}", string(marshal))
			})
			r.Success(ctx, x)
		})

		t.Run("data is nil", func(t *testing.T) {

			ctx.EXPECT().JSON(gomock.Any(), gomock.Any()).Do(func(code int, obj any) {
				marshal, err := json.Marshal(obj)
				assert.Nil(t, err)

				assert.Equal(t, "{\"code\":0}", string(marshal))
			})
			r.Success(ctx, nil)
		})

		t.Run("data is not nil", func(t *testing.T) {
			type Data struct {
				X int `json:"x"`
			}

			var x = Data{X: 128}

			ctx.EXPECT().JSON(gomock.Any(), gomock.Any()).Do(func(code int, obj any) {
				marshal, err := json.Marshal(obj)
				assert.Nil(t, err)

				assert.Equal(t, "{\"code\":0,\"data\":{\"x\":128}}", string(marshal))
			})
			r.Success(ctx, x)
		})
	})

}

func Test_responser_Failed(t *testing.T) {
	t.Run("returnWrappedData=false", func(t *testing.T) {
		controller := gomock.NewController(t)
		defer controller.Finish()

		r := responser{
			wrappedDataFunc: wrapFunc,
		}

		ctx := NewMockXContext(controller)

		t.Run("oErr is InnerError", func(t *testing.T) {
			x := NewInnerError("error", 1)
			ctx.EXPECT().String(gomock.Any(), gomock.Any()).Do(func(code int, format string, values ...any) {
				assert.Equal(t, http.StatusInternalServerError, code)
				assert.Equal(t, "error", format)
			})
			r.Failed(ctx, x)
		})

		t.Run("oErr is nil", func(t *testing.T) {
			ctx.EXPECT().String(gomock.Any(), gomock.Any()).Do(func(code int, format string, values ...any) {
				assert.Equal(t, http.StatusBadRequest, code)
				assert.Equal(t, "", format)
			})
			r.Failed(ctx, nil)
		})

		t.Run("oErr is not nil", func(t *testing.T) {
			x := NewBusinessError("test", 1, "test-data")

			ctx.EXPECT().String(gomock.Any(), gomock.Any()).Do(func(code int, format string, values ...any) {
				assert.Equal(t, http.StatusBadRequest, code)
				assert.Equal(t, "GoneError(code=1):test", format)
			})
			r.Failed(ctx, x)
		})
	})

	t.Run("returnWrappedData=false", func(t *testing.T) {
		controller := gomock.NewController(t)
		defer controller.Finish()

		r := responser{
			wrappedDataFunc:   wrapFunc,
			returnWrappedData: true,
		}

		ctx := NewMockXContext(controller)

		t.Run("oErr is BusinessError", func(t *testing.T) {
			x := NewBusinessError("test", 1, "test-data")

			ctx.EXPECT().JSON(gomock.Any(), gomock.Any()).Do(func(code int, obj any) {
				assert.Equal(t, http.StatusOK, code)
				marshal, err := json.Marshal(obj)
				assert.Nil(t, err)

				assert.Equal(t, "{\"code\":1,\"msg\":\"test\",\"data\":\"test-data\"}", string(marshal))
			})
			r.Failed(ctx, x)
		})

		t.Run("oErr is nil", func(t *testing.T) {
			ctx.EXPECT().JSON(gomock.Any(), gomock.Any()).Do(func(code int, obj any) {
				assert.Equal(t, http.StatusBadRequest, code)

				marshal, err := json.Marshal(obj)
				assert.Nil(t, err)

				assert.Equal(t, "{\"code\":0}", string(marshal))
			})
			r.Failed(ctx, nil)
		})

		t.Run("oErr is InnerError", func(t *testing.T) {
			var x = NewInnerError("error", 1)

			ctx.EXPECT().JSON(gomock.Any(), gomock.Any()).Do(func(code int, obj any) {
				assert.Equal(t, http.StatusInternalServerError, code)

				marshal, err := json.Marshal(obj)
				assert.Nil(t, err)

				assert.Equal(t, "{\"code\":1,\"msg\":\"Internal Server Error\"}", string(marshal))
			})
			r.Failed(ctx, x)
		})

		t.Run("oErr is Error", func(t *testing.T) {
			var x = NewParameterError("test", 1)

			ctx.EXPECT().JSON(gomock.Any(), gomock.Any()).Do(func(code int, obj any) {
				assert.Equal(t, http.StatusBadRequest, code)

				marshal, err := json.Marshal(obj)
				assert.Nil(t, err)

				assert.Equal(t, "{\"code\":1,\"msg\":\"test\"}", string(marshal))
			})
			r.Failed(ctx, x)
		})

		t.Run("oErr is other", func(t *testing.T) {
			ctx.EXPECT().JSON(gomock.Any(), gomock.Any()).Do(func(code int, obj any) {
				assert.Equal(t, http.StatusInternalServerError, code)
				marshal, err := json.Marshal(obj)
				assert.Nil(t, err)

				assert.Equal(t, "{\"code\":500,\"msg\":\"Internal Server Error\"}", string(marshal))
			})
			r.Failed(ctx, errors.New("test"))
		})
	})

}

func Test_responser_ProcessResults(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	r := responser{
		wrappedDataFunc:   wrapFunc,
		returnWrappedData: true,
	}

	ctx := NewMockXContext(controller)
	writer := NewMockResponseWriter(controller)

	ctx.EXPECT().JSON(gomock.Any(), gomock.Any()).AnyTimes()

	t.Run("last=false", func(t *testing.T) {
		writer.EXPECT().Written().Return(false).AnyTimes()

		t.Run("results is [nil, error]", func(t *testing.T) {

			r.ProcessResults(ctx, writer, false, "test", nil, errors.New("test"))
		})

		t.Run("results is [struct]", func(t *testing.T) {
			r.ProcessResults(ctx, writer, false, "test",
				struct {
					X int
				}{X: 100},
			)
		})

		t.Run("results is [io.Reader]", func(t *testing.T) {
			writer.EXPECT().WriteString(gomock.Any()).Return(0, nil).AnyTimes()
			r.ProcessResults(ctx, writer, false, "test", strings.NewReader("this is a test"))
		})

		t.Run("results is [chan any]", func(t *testing.T) {
			ch := make(chan any)
			go func() {
				defer close(ch)
				ch <- "test"
				ch <- errors.New("errr")
			}()

			writer.EXPECT().Header().Return(http.Header{}).AnyTimes()
			writer.EXPECT().Flush().AnyTimes()
			writer.EXPECT().WriteString(gomock.Any()).Return(0, nil).AnyTimes()
			writer.EXPECT().CloseNotify().AnyTimes()

			r.ProcessResults(ctx, writer, false, "test", ch)
		})

	})

	t.Run("last=true", func(t *testing.T) {
		writer.EXPECT().Written().Return(true).AnyTimes()
		t.Run("results is [nil, error]", func(t *testing.T) {
			r.ProcessResults(ctx, writer, true, "test", "test")
		})
	})

	t.Run("WriteString error", func(t *testing.T) {
		ch := make(chan any)
		go func() {
			defer close(ch)
			ch <- "test"
			ch <- errors.New("errr")
		}()

		writer.EXPECT().Header().Return(http.Header{}).AnyTimes()
		writer.EXPECT().Flush().AnyTimes()
		writer.EXPECT().WriteString(gomock.Any()).Return(0, errors.New("error")).AnyTimes()
		writer.EXPECT().CloseNotify().AnyTimes()
		writer.EXPECT().Written().Return(false).AnyTimes()

		r.ProcessResults(ctx, writer, false, "test", ch)
	})
}

func Test_responser_SetWrappedDataFunc(t *testing.T) {
	r := responser{
		returnWrappedData: true,
	}
	r.SetWrappedDataFunc(wrapFunc)
	assert.NotNil(t, r.wrappedDataFunc)
}
