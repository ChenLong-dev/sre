package gin

import (
	"database/sql/driver"
	"reflect"
	"sync"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"gitlab.shanhai.int/sre/library/base/null"
)

type defaultV10Validator struct {
	// 用于懒加载
	once sync.Once
	// 校验引擎
	validate *validator.Validate
}

var _ binding.StructValidator = &defaultV10Validator{}

// 校验结构体
func (v *defaultV10Validator) ValidateStruct(obj interface{}) error {
	if kindOfData(obj) == reflect.Struct {
		v.lazyInit()
		if err := v.validate.Struct(obj); err != nil {
			return error(err)
		}
	}
	return nil
}

// 获取校验引擎
func (v *defaultV10Validator) Engine() interface{} {
	v.lazyInit()
	return v.validate
}

// 懒加载
func (v *defaultV10Validator) lazyInit() {
	v.once.Do(func() {
		v.validate = validator.New()
		v.validate.SetTagName("binding")
		v.validate.RegisterCustomTypeFunc(func(field reflect.Value) interface{} {
			if valuer, ok := field.Interface().(driver.Valuer); ok {
				val, err := valuer.Value()
				if err == nil {
					return val
				}
			}
			return nil
		}, null.String{}, null.Int64{}, null.Bool{}, null.Float{}, null.Int{}, null.Time{})
	})
}

// 判断data类型
func kindOfData(data interface{}) reflect.Kind {
	value := reflect.ValueOf(data)
	valueType := value.Kind()
	if valueType == reflect.Ptr {
		valueType = value.Elem().Kind()
	}
	return valueType
}

// 新建V10校验器
// 支持null类型
func NewV10Validator() *defaultV10Validator {
	return new(defaultV10Validator)
}
