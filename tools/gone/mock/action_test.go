package mock

import (
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
	"testing"
)

func TestAction(t *testing.T) {
	t.Run("bad ags", func(t *testing.T) {
		app := cli.App{
			Commands: []*cli.Command{
				CreateCommand(),
			},
		}

		err := app.Run([]string{"", "mock"})
		assert.Error(t, err)
	})

	t.Run("good args", func(t *testing.T) {
		app := cli.App{
			Commands: []*cli.Command{
				CreateCommand(),
			},
		}

		err := app.Run([]string{"", "mock", "-f", "testdata/testInterface.go", "-o", "testdata/mock/testInterface.go"})
		assert.Nil(t, err)
	})
}
