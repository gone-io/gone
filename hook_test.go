package gone

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBeforeStartProvider_Provide(t *testing.T) {
	// 准备测试数据
	preparer := &Preparer{}
	provider := &BeforeStartProvider{
		preparer: preparer,
	}

	// 执行测试
	hook, err := provider.Provide()

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, hook)
	var beforeStart = BeforeStart(preparer.beforeStart)

	assert.Equal(t, reflect.ValueOf(hook).Pointer(), reflect.ValueOf(beforeStart).Pointer())
}

func TestAfterStartProvider_Provide(t *testing.T) {
	// 准备测试数据
	preparer := &Preparer{}
	provider := &AfterStartProvider{
		preparer: preparer,
	}

	// 执行测试
	hook, err := provider.Provide()

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, hook)
	var fn = AfterStart(preparer.afterStart)
	assert.Equal(t, reflect.ValueOf(hook).Pointer(), reflect.ValueOf(fn).Pointer())
}

func TestBeforeStopProvider_Provide(t *testing.T) {
	// 准备测试数据
	preparer := &Preparer{}
	provider := &BeforeStopProvider{
		preparer: preparer,
	}

	// 执行测试
	hook, err := provider.Provide()

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, hook)
	var fn = BeforeStop(preparer.beforeStop)
	assert.Equal(t, reflect.ValueOf(hook).Pointer(), reflect.ValueOf(fn).Pointer())
}

func TestAfterStopProvider_Provide(t *testing.T) {
	// 准备测试数据
	preparer := &Preparer{}
	provider := &AfterStopProvider{
		preparer: preparer,
	}

	// 执行测试
	hook, err := provider.Provide()

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, hook)
	var fn = AfterStop(preparer.afterStop)
	assert.Equal(t, reflect.ValueOf(hook).Pointer(), reflect.ValueOf(fn).Pointer())
}
