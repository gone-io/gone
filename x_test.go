package gone_test

import (
	"fmt"
	"github.com/gone-io/gone"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"unsafe"
)

func getStructPtrUnExportedField(source interface{}, fieldName string) reflect.Value {
	// 获取非导出字段反射对象
	v := reflect.ValueOf(source).Elem().FieldByName(fieldName)
	t, b := reflect.TypeOf(source).Elem().FieldByName(fieldName)
	if b {
		tag := t.Tag
		print(tag)
	}

	// 构建指向该字段的可寻址（addressable）反射对象
	return reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem()
}

func setStructPtrUnExportedStrField(source interface{}, fieldName string, fieldVal interface{}) (err error) {
	v := getStructPtrUnExportedField(source, fieldName)
	rv := reflect.ValueOf(fieldVal)
	if v.Kind() != rv.Kind() {
		return fmt.Errorf("invalid kind: expected kind %v, got kind: %v", v.Kind(), rv.Kind())
	}
	// 修改非导出字段值
	v.Set(rv)
	return nil
}

func Test_SetPrivateProperty(t *testing.T) {
	var a = new(gone.A)
	var b = new(gone.B)

	err := setStructPtrUnExportedStrField(a, "a", int(4))
	assert.Nil(t, err)

	err = setStructPtrUnExportedStrField(b, "a", *a)
	assert.Nil(t, err)
}
