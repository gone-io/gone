package gin

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/internal/json"
	"io"
)

func NewSSE(writer gin.ResponseWriter) SSE {
	return &Sse{Writer: writer}
}

type SSE interface {
	Start()
	Write(delta any) error
	End() error
	WriteError(err gone.Error) error
}

type Sse struct {
	Writer gin.ResponseWriter
}

func (s *Sse) Start() {
	s.Writer.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	s.Writer.Header().Set("Cache-Control", "no-cache")
	s.Writer.Header().Set("Connection", "keep-alive")
	s.Writer.Header().Set("X-Accel-Buffering", "no")
	s.Writer.Flush()
}

func (s *Sse) Write(delta any) error {
	jsonStr, err := json.Marshal(delta)
	if err != nil {
		return err
	}

	_, err = io.WriteString(s.Writer, "event: data\n")
	if err != nil {
		return err
	}
	_, err = io.WriteString(s.Writer, fmt.Sprintf("data: %s\n\n", jsonStr))
	if err != nil {
		return err
	}
	s.Writer.Flush()
	return nil
}

func (s *Sse) End() error {
	_, err := io.WriteString(s.Writer, "event: done\n")
	if err != nil {
		return err
	}
	s.Writer.Flush()
	s.Writer.CloseNotify()
	return nil
}
func (s *Sse) WriteError(err gone.Error) error {
	var x struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	x.Code = err.Code()
	x.Msg = err.Error()
	return s.Write(x)
}
