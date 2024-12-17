package gone

import (
	"errors"
	"reflect"
	"testing"
)

// Mock providers for testing
type SimpleProvider struct {
	returnVal string
	returnErr error
}

func (p *SimpleProvider) Provide() (string, error) {
	return p.returnVal, p.returnErr
}

type ConfigurableProvider struct {
	returnVal string
	returnErr error
}

func (p *ConfigurableProvider) Provide(conf string) (string, error) {
	if conf == "error" {
		return "", errors.New("configured error")
	}
	return p.returnVal + "-" + conf, p.returnErr
}

// Invalid providers for testing
type InvalidProvider1 struct{}

func (p InvalidProvider1) Provide() (string, error) { // Not a pointer receiver
	return "", nil
}

type InvalidProvider2 struct{}

func (p *InvalidProvider2) Provide(a, b string) (string, error) { // Wrong number of parameters
	return "", nil
}

type InvalidProvider3 struct{}

func (p *InvalidProvider3) Provide() string { // Wrong return types
	return ""
}

func TestTryWrapGonerToProvider(t *testing.T) {
	tests := []struct {
		name       string
		goner      any
		wantNil    bool
		wantType   reflect.Type
		hasParam   bool
		wantErrMsg string
	}{
		{
			name: "Valid simple provider",
			goner: &SimpleProvider{
				returnVal: "test",
				returnErr: nil,
			},
			wantNil:  false,
			wantType: reflect.TypeOf(""),
			hasParam: false,
		},
		{
			name: "Valid configurable provider",
			goner: &ConfigurableProvider{
				returnVal: "test",
				returnErr: nil,
			},
			wantNil:  false,
			wantType: reflect.TypeOf(""),
			hasParam: true,
		},
		{
			name:     "Non-pointer receiver",
			goner:    InvalidProvider1{},
			wantNil:  true,
			wantType: nil,
			hasParam: false,
		},
		{
			name:     "Wrong parameter count",
			goner:    &InvalidProvider2{},
			wantNil:  true,
			wantType: nil,
			hasParam: false,
		},
		{
			name:     "Wrong return types",
			goner:    &InvalidProvider3{},
			wantNil:  true,
			wantType: nil,
			hasParam: false,
		},
		{
			name:     "Nil input",
			goner:    nil,
			wantNil:  true,
			wantType: nil,
			hasParam: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := tryWrapGonerToProvider(tt.goner)

			if tt.wantNil {
				if provider != nil {
					t.Errorf("tryWrapGonerToProvider() = %v, want nil", provider)
				}
				return
			}

			if provider == nil {
				t.Fatal("tryWrapGonerToProvider() = nil, want non-nil")
			}

			if provider.hasParameter != tt.hasParam {
				t.Errorf("provider.hasParameter = %v, want %v", provider.hasParameter, tt.hasParam)
			}

			if !reflect.DeepEqual(provider.t, tt.wantType) {
				t.Errorf("provider.t = %v, want %v", provider.t, tt.wantType)
			}
		})
	}
}

func TestWrapProvider_Provide(t *testing.T) {
	tests := []struct {
		name      string
		provider  *wrapProvider
		conf      string
		want      any
		wantError bool
	}{
		{
			name: "Simple provider success",
			provider: &wrapProvider{
				value: &SimpleProvider{
					returnVal: "test",
					returnErr: nil,
				},
				hasParameter: false,
				t:            reflect.TypeOf(""),
			},
			conf:      "",
			want:      "test",
			wantError: false,
		},
		{
			name: "Simple provider error",
			provider: &wrapProvider{
				value: &SimpleProvider{
					returnVal: "",
					returnErr: errors.New("test error"),
				},
				hasParameter: false,
				t:            reflect.TypeOf(""),
			},
			conf:      "",
			want:      "",
			wantError: true,
		},
		{
			name: "Configurable provider success",
			provider: &wrapProvider{
				value: &ConfigurableProvider{
					returnVal: "test",
					returnErr: nil,
				},
				hasParameter: true,
				t:            reflect.TypeOf(""),
			},
			conf:      "config",
			want:      "test-config",
			wantError: false,
		},
		{
			name: "Configurable provider error",
			provider: &wrapProvider{
				value: &ConfigurableProvider{
					returnVal: "test",
					returnErr: nil,
				},
				hasParameter: true,
				t:            reflect.TypeOf(""),
			},
			conf:      "error",
			want:      "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.provider.Provide(tt.conf)

			if (err != nil) != tt.wantError {
				t.Errorf("wrapProvider.Provide() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("wrapProvider.Provide() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWrapProvider_Type(t *testing.T) {
	tests := []struct {
		name     string
		provider *wrapProvider
		want     reflect.Type
	}{
		{
			name: "String type",
			provider: &wrapProvider{
				t: reflect.TypeOf(""),
			},
			want: reflect.TypeOf(""),
		},
		{
			name: "Int type",
			provider: &wrapProvider{
				t: reflect.TypeOf(0),
			},
			want: reflect.TypeOf(0),
		},
		{
			name: "Struct type",
			provider: &wrapProvider{
				t: reflect.TypeOf(struct{}{}),
			},
			want: reflect.TypeOf(struct{}{}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.provider.Type(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("wrapProvider.Type() = %v, want %v", got, tt.want)
			}
		})
	}
}
