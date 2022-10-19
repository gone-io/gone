package worker

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/logrus"
	"github.com/gone-io/gone/goner/tracer"
	"time"
)

const IdTimerWorker = "worker-timer"

//go:gone
func NewTimerWorker() (gone.Goner, gone.GonerId) {
	return &TimerWorker{}, IdTimerWorker
}

type TimerWorker struct {
	gone.Flag

	logrus.Logger `gone:"gone-logger"`
	worker        PrintWorker   `gone:"*"`
	tracer        tracer.Tracer `gone:"gone-tracer"`

	Ttl int `gone:"config,example.app.print.ttl"`
}

func (w *TimerWorker) AfterRevive(gone.Cemetery, gone.Tomb) gone.ReviveAfterError {
	w.tracer.SetTraceId("", func() {
		w.tracer.Go(func() {
			w.Printf("I will print a log every %d second", w.Ttl)
			for true {
				<-time.After(time.Duration(w.Ttl) * time.Second)
				w.worker.Print()
			}
		})
	})
	return nil
}
