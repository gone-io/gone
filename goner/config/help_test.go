package config

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		before func()
		name   string
		want   string
		after  func()
	}{
		{
			name: "read value from env",
			before: func() {
				_ = os.Setenv(EEnv, "dev")
			},
			after: func() {
				_ = os.Unsetenv(EEnv)
			},
			want: "dev",
		}, {
			name:   "read value from default",
			want:   "local",
			after:  func() {},
			before: func() {},
		},
		{
			name: "read value from flag",
			before: func() {
				os.Args = append(os.Args, "-env=test")
			},
			want: "test",
			after: func() {
				for i, v := range os.Args {
					if v == "-env=test" {
						os.Args = append(os.Args[:i], os.Args[i+1:]...)
					}
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before()
			assert.Equalf(t, tt.want, GetEnv(), "GetEnv()")
			tt.after()
		})
	}
}

func TestGetConfSettings(t *testing.T) {
	os.Args = append(os.Args, "-conf=x-config")

	configs := GetConfSettings(true)
	assert.Equal(t, 20, len(configs))
	assert.Equal(t, "x-config", configs[19].ConfigPath)
}
