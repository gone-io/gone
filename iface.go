package gone

import (
	"fmt"
	"reflect"
	"strings"
)

// Flag is a marker struct used to identify components that can be managed by the gone framework.
// Embedding this struct in another struct indicates that it can be used with gone's dependency injection.
type Flag struct{}

func (g *Flag) goneFlag() {}

type actionType int8

func (t actionType) String() string {
	switch t {
	case fillAction:
		return "fill fields"
	case initAction:
		return "initialize"
	default:
		return "unknown"
	}
}

type dependency struct {
	coffin *coffin
	action actionType
}

func (d dependency) String() string {
	var name string
	if d.coffin.name != "" {
		name = fmt.Sprintf("%q", d.coffin.name)
	} else {
		name = fmt.Sprintf("%q", GetTypeName(reflect.TypeOf(d.coffin.goner)))
	}
	return fmt.Sprintf("<%s of %s>", d.action.String(), name)
}

const (
	fillAction          actionType = 1
	initAction          actionType = 2
	goneTag                        = "gone"
	DefaultProviderName            = "core-provider"
	optionTag                      = "option"
	allowNil                       = "allowNil"
	lazy                           = "lazy"
)

func filedHasOption(filed *reflect.StructField, tagName string, optionName string) bool {
	value, ok := filed.Tag.Lookup(tagName)
	if !ok || value == "" {
		return false
	}
	split := strings.Split(value, ",")
	for _, v := range split {
		if v == optionName {
			return true
		}
	}
	return false
}

func isAllowNilField(filed *reflect.StructField) bool {
	return filedHasOption(filed, optionTag, allowNil)
}
func isLazyField(filed *reflect.StructField) bool {
	return filedHasOption(filed, optionTag, lazy)
}

// FuncInjectHook is a function type used for customizing parameter injection in functions.
// Parameters:
//   - pt: The type of parameter being injected
//   - i: The index of the parameter in the function signature
//   - injected: Whether the parameter has already been injected
//
// Returns any value that should be used as the injected parameter, or nil to continue with default injection
type FuncInjectHook func(pt reflect.Type, i int, injected bool) any

//go:generate mockgen  -source=./iface.go    -package=gone -destination=iface_mock.go
//go:generate mockgen -source=./interface.go -package=gone -destination=interface_mock.go
//go:generate mockgen -source=./logger.go    -package=gone -destination=./logger_mock.go
//go:generate mockgen -source=./config.go    -package=gone -destination=./config_mock.go

type iKeeper interface {
	load(goner Goner, options ...Option) error
	getAllCoffins() []*coffin
	getByTypeAndPattern(t reflect.Type, pattern string) []*coffin
	selectOneCoffin(t reflect.Type, pattern string, warn func()) (depCo *coffin)
	getByName(name string) *coffin
}

type iDependenceAnalyzer interface {
	analyzerFieldDependencies(
		field reflect.StructField, coName string,
		process func(asSlice, byName bool, extend string, coffins ...*coffin) error,
	) error

	checkCircularDepsAndGetBestInitOrder() (circularDeps []dependency, initOrder []dependency, err error)
}

type iInstaller interface {
	safeFillOne(c *coffin) error
	safeInitOne(c *coffin) error

	analyzerFieldDependencies(
		field reflect.StructField, coName string,
		process func(asSlice, byName bool, extend string, coffins ...*coffin) error,
	) error

	injectField(
		asSlice, byName bool, extend string, depCoffins []*coffin,
		field reflect.StructField, v reflect.Value, coName string,
	) error
}
