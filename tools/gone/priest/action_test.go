package priest

import (
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
	"testing"
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
		app := cli.App{
			Commands: []*cli.Command{
				CreateCommand(),
			},
		}

		err := app.Run([]string{"", "priest", "-s", "testdata/x", "-p", "x", "-f", "priest", "-o", "testdata/x/priest.go", "--stat"})
		assert.Nil(t, err)
	})
}
