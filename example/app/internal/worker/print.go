package worker

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/logrus"
)

type PrintWorker interface {
	gone.Goner
	Print()
	GetContent() string
}

//go:gone
func NewPrintWorker() PrintWorker {
	return &printWorker{}
}

type printWorker struct {
	gone.Flag
	content       string `gone:"config,example.app.print.content"`
	logrus.Logger `gone:"gone-logger"`
}

func (w *printWorker) Print() {
	w.Println(w.content)
}

func (w *printWorker) GetContent() string {
	return w.content
}
