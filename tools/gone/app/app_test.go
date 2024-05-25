package app

import (
	"embed"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func deleteFilesInDirectory(dir string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %v", err)
	}

	for _, file := range files {
		if !file.IsDir() {
			err := os.Remove(filepath.Join(dir, file.Name()))
			if err != nil {
				return fmt.Errorf("failed to delete file %s: %v", file.Name(), err)
			}
		}
	}
	return nil
}

//go:embed testdata/from/*
var testF embed.FS

func Test_copyAndReplace(t *testing.T) {
	_ = deleteFilesInDirectory("testdata/to/")
	err := copyAndReplace(testF, "testdata/from", "testdata/to/", map[string]string{
		"test": "x-test",
	})
	assert.Nil(t, err)
}

func Test_paramsProcess(t *testing.T) {
	tpl, m, app, err := paramsProcess("", "", "")
	assert.Error(t, err)

	tpl, m, app, err = paramsProcess("web", "", "")
	assert.Nil(t, err)
	assert.Equal(t, "web", tpl)
	assert.Equal(t, "demo", m)
	assert.Equal(t, "demo", app)
}
