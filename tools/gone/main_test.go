package main

import (
	"testing"
)

func Test_run(t *testing.T) {
	run("-h")

	run("", "priest")
}
