package use_case

import (
	"github.com/gone-io/gone/v2"
	"testing"
)

type app struct {
	gone.Flag
	log gone.Logger `gone:"*"`
}

func TestUseLogger(t *testing.T) {
	gone.
		Load(&app{}).
		Run(func(app *app) {
			app.log.Infof("hello world")
			app.log.Errorf("hello world")
			app.log.Warnf("hello world")
			app.log.Debugf("hello world")
		})
}
