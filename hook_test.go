package gone

import (
	"reflect"
	"testing"
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
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if hook == nil {
		t.Error("Expected hook to not be nil")
	}
	var beforeStart = BeforeStart(preparer.beforeStart)

	if reflect.ValueOf(hook).Pointer() != reflect.ValueOf(beforeStart).Pointer() {
		t.Error("Hook functions do not match")
	}
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
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if hook == nil {
		t.Error("Expected hook to not be nil")
	}
	var fn = AfterStart(preparer.afterStart)
	if reflect.ValueOf(hook).Pointer() != reflect.ValueOf(fn).Pointer() {
		t.Error("Hook functions do not match")
	}
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
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if hook == nil {
		t.Error("Expected hook to not be nil")
	}
	var fn = BeforeStop(preparer.beforeStop)
	if reflect.ValueOf(hook).Pointer() != reflect.ValueOf(fn).Pointer() {
		t.Error("Hook functions do not match")
	}
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
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if hook == nil {
		t.Error("Expected hook to not be nil")
	}
	var fn = AfterStop(preparer.afterStop)
	if reflect.ValueOf(hook).Pointer() != reflect.ValueOf(fn).Pointer() {
		t.Error("Hook functions do not match")
	}
}
