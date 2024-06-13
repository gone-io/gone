package priest

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
	"os"
	"testing"
	"time"
)

func TestAction(t *testing.T) {
	t.Run("bad args", func(t *testing.T) {
		app := cli.App{
			Commands: []*cli.Command{
				CreateCommand(),
			},
		}

		err := app.Run([]string{"", "priest"})
		assert.Error(t, err)
	})

	t.Run("good args", func(t *testing.T) {
		_ = os.Remove("testdata/x/goner-new.go")
		app := cli.App{
			Commands: []*cli.Command{
				CreateCommand(),
			},
		}

		ch := getWatchDoneChannel()
		go func() {
			time.Sleep(1 * time.Second)
			file, _ := os.OpenFile("testdata/x/goner-new.go", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
			defer func() {
				_ = file.Close()
				time.Sleep(1 * time.Second)
				close(ch)
			}()
			_, _ = file.WriteString(`package x

import "github.com/gone-io/gone"

//go:gone
func New2() gone.Goner {
	return &goner{}
}`)
		}()

		err := app.Run([]string{"", "priest",
			"-s", "testdata/x",
			"-p", "x",
			"-f", "priest",
			"-o", "testdata/x/priest.go",
			"--stat",
			"-w",
		})
		assert.Nil(t, err)

		file, err := os.ReadFile("testdata/x/priest.go")
		assert.Nil(t, err)
		assert.True(t, bytes.Contains(file, []byte("cemetery.Bury(New2())")))
	})
}
