package gin

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/gone-io/gone"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestSSE(t *testing.T) {
	t.Run("suc", func(t *testing.T) {
		controller := gomock.NewController(t)
		defer controller.Finish()

		writer := NewMockResponseWriter(controller)
		writer.EXPECT().Header().Return(http.Header{}).AnyTimes()
		writer.EXPECT().Flush().AnyTimes()
		writer.EXPECT().WriteString(gomock.Any()).Return(100, nil).AnyTimes()
		writer.EXPECT().CloseNotify().Return(nil)

		sse := NewSSE(writer)

		sse.Start()
		err := sse.Write(map[string]any{
			"key": "value",
		})

		assert.Nil(t, err)

		err = sse.WriteError(gone.NewError(100, "error"))
		assert.Nil(t, err)

		err = sse.End()
		assert.Nil(t, err)
	})

	t.Run("json Marshal error", func(t *testing.T) {
		controller := gomock.NewController(t)
		defer controller.Finish()

		writer := NewMockResponseWriter(controller)
		writer.EXPECT().Header().Return(http.Header{}).AnyTimes()
		writer.EXPECT().WriteString(gomock.Any()).Return(100, nil).AnyTimes()
		writer.EXPECT().Flush().AnyTimes()

		sse := NewSSE(writer)

		sse.Start()
		err := sse.Write(map[string]any{
			"key": func() {},
		})

		assert.Error(t, err)
	})

	t.Run("write err", func(t *testing.T) {
		controller := gomock.NewController(t)
		defer controller.Finish()

		writer := NewMockResponseWriter(controller)
		writer.EXPECT().Header().Return(http.Header{}).AnyTimes()
		writer.EXPECT().Flush().AnyTimes()
		writer.EXPECT().WriteString(gomock.Any()).Return(0, errors.New("error")).AnyTimes()

		sse := NewSSE(writer)

		sse.Start()
		err := sse.Write(map[string]any{
			"key": "value",
		})

		assert.Error(t, err)
	})

	t.Run("write err", func(t *testing.T) {
		controller := gomock.NewController(t)
		defer controller.Finish()

		writer := NewMockResponseWriter(controller)
		writer.EXPECT().Header().Return(http.Header{}).AnyTimes()
		writer.EXPECT().Flush().AnyTimes()
		writer.EXPECT().WriteString(gomock.Any()).Return(100, nil)
		writer.EXPECT().WriteString(gomock.Any()).Return(0, errors.New("error"))

		sse := NewSSE(writer)

		sse.Start()
		err := sse.Write(map[string]any{
			"key": "value",
		})

		assert.Error(t, err)
	})

	t.Run("End write err", func(t *testing.T) {
		controller := gomock.NewController(t)
		defer controller.Finish()

		writer := NewMockResponseWriter(controller)
		writer.EXPECT().Header().Return(http.Header{}).AnyTimes()
		writer.EXPECT().Flush().AnyTimes()
		writer.EXPECT().WriteString(gomock.Any()).Return(0, errors.New("error")).AnyTimes()

		sse := NewSSE(writer)

		sse.Start()
		err := sse.End()

		assert.Error(t, err)
	})
}
