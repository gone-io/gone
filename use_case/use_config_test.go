package use_case

import (
	"github.com/gone-io/gone/v2"
	"os"
	"testing"
)

type useConfig struct {
	gone.Flag
	goneVersion string `gone:"config,gone-version"` //框架在启动时，会自动加载配置，并注入到goRoot字段中。
}

func TestUseConfig(t *testing.T) {
	os.Setenv("GONE_GONE-VERSION", "v2.0.0")

	gone.
		Load(&useConfig{}).
		Run(func(c *useConfig) {
			println("goRoot:", c.goneVersion)
			if c.goneVersion != "v2.0.0" {
				t.Fatal("配置注入失败")
			}
		})
}
