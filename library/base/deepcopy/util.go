package deepcopy

import (
	"database/sql/driver"
	"reflect"
	"time"
)

// 获取对应结构体的方法名
func getMethodNames(instance interface{}) []string {
	var methods []string

	t := reflect.TypeOf(instance)
	for i := 0; i < t.NumMethod(); i++ {
		methods = append(methods, t.Method(i).Name)
	}

	return methods
}

// 是否可以转换为driver.Valuer类型
func canConvertToValuerType(t reflect.Type) bool {
	return t.ConvertibleTo(reflect.TypeOf((*driver.Valuer)(nil)).Elem())
}

// 是否可以转换为time.Time类型
func canConvertToTimeType(t reflect.Type) bool {
	return t.ConvertibleTo(reflect.TypeOf((*time.Time)(nil)).Elem())
}
