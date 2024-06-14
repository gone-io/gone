package gone_zap

import (
	"github.com/gone-io/gone"
	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

type traceEncoder struct {
	zapcore.Encoder
	tracer gone.Tracer
}

func (e *traceEncoder) EncodeEntry(entry zapcore.Entry, fields []Field) (*buffer.Buffer, error) {
	traceId := e.tracer.GetTraceId()
	if traceId != "" {
		fields = append(fields, zap.String("traceId", traceId))
	}
	return e.Encoder.EncodeEntry(entry, fields)
}

func NewTraceEncoder(encoder zapcore.Encoder, tracer gone.Tracer) zapcore.Encoder {
	return &traceEncoder{
		Encoder: encoder,
		tracer:  tracer,
	}
}
