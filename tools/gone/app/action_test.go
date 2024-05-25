package app

import (
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
	"testing"
)

func TestAction(t *testing.T) {
	app := cli.App{
		Commands: []*cli.Command{
			CreateCommand(),
		},
	}

	err := app.Run([]string{"", "create"})
	assert.Nil(t, err)
}
