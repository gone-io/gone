package priest

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestScanDir(t *testing.T) {
	t.Run("filepath not existed", func(t *testing.T) {
		_, err := ScanDir("xxxx/xxxxx", "", "")
		assert.Error(t, err)
	})

	t.Run("filepath is a file", func(t *testing.T) {
		_, err := ScanDir("testdata/x/sub/.keep", "", "")
		assert.Error(t, err)
	})

	t.Run("filepath is empty", func(t *testing.T) {
		_, err := ScanDir("testdata/x/sub", "", "")
		assert.Nil(t, err)
	})
}

func TestPkg_generateFuncContent(t *testing.T) {
	a := Pkg{}
	content := a.generateFuncContent(true)
	assert.Equal(t, "", content)

	a.Name = "yyy"
	a.PkgPath = "zzz"
	importContent := a.genImportContent()
	assert.Equal(t, "    yyy \"zzz\"", importContent)
}
