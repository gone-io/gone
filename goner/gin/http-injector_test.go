package gin

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_HttpInject(t *testing.T) {
	gone.
		Prepare(func(cemetery gone.Cemetery) error {
			_ = Priest(cemetery)

			var x struct {
				gone.Flag
			}

			_ = cemetery.ReplaceBury(&x, gone.IdGoneGin)
			return nil
		}).
		AfterStart(func(in struct {
			p         *proxy        `gone:"gone-gin-proxy"`
			processor *sysProcessor `gone:"gone-gin-processor"`

			log logrus.Logger `gone:"gone-logger"`
		}) {

			context := Context{}
			i := 0

			funcs := in.p.Proxy(in.processor.trace, func(arg struct {
				page     int    `gone:"http,query=page"`
				cookX    int    `gone:"http,cookie=x"`
				headerY  string `gone:"http,header=y"`
				token    string `gone:"http,auth=Bearer"`
				formData string `gone:"http,form=data"`

				host    string   `gone:"http,host"`
				url     string   `gone:"http,url"`
				path    string   `gone:"http,path"`
				query   string   `gone:"http,query"`
				data    string   `gone:"http,body"`
				context *Context `gone:"http,context"`

				log logrus.Logger `gone:"gone-logger"`
			}) (int, error) {
				i++
				assert.Equal(t, 5, i)

				assert.Equal(t, in.log, arg.log)
				assert.NotNil(t, in.log)
				assert.Equal(t, &context, arg.context.Context)

				return 0, nil
			})

			funcs[0](context.Context)

			assert.Equal(t, 5, i)
		}).
		Run()
}
