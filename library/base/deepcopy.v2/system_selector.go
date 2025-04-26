package deepcopy

import (
	"github.com/pkg/errors"
	"reflect"
	"strconv"
)

// 选择器方法
type SelectorFunc func(dc *DeepCopier, src ProcessorValue, dst ProcessorValue) ProcessorValue

// default标签选择器
var DefaultTagSelector = func(dc *DeepCopier, src ProcessorValue, dst ProcessorValue) ProcessorValue {
	if src.GetValueType() != ProcessorValueSingleType || dst.GetValueType() != ProcessorValueStructFieldType {
		return dc.NextSelector(src, dst)
	}

	defaultName, ok := dst.GetStructFieldTag(DefaultMethodName)
	if !ok {
		return dc.NextSelector(src, dst)
	}

	dstParent := dst.GetParentStruct()

	// 判断结构体函数函数
	dstParentPtr := dstParent
	if dstParent.Type().Kind() == reflect.Struct {
		dstParentPtr = reflect.New(dstParent.Type())
		dstParentPtr.Elem().Set(dstParent)
	}
	if dstParentPtr.IsValid() && dstParentPtr.Type().Kind() == reflect.Ptr {
		for i := 0; i < dstParentPtr.Type().NumMethod(); i++ {
			dstMethodValue := dstParentPtr.Method(i)

			if !dstMethodValue.IsValid() {
				continue
			}

			if dstParentPtr.Type().Method(i).Name != defaultName {
				continue
			}

			result := dstMethodValue.Call([]reflect.Value{reflect.ValueOf(dc.args)})
			if len(result) < 1 {
				dc.SetError(errors.Errorf("`default` tag method:%s return empty", defaultName))
				return NewDefaultProcessorValue()
			} else if len(result) > 1 {
				err, ok := result[len(result)-1].Interface().(error)
				if ok && err != nil {
					dc.SetError(err)
					return NewDefaultProcessorValue()
				}
			}

			return NewSingleProcessorValue(result[0])
		}
	}

	// 判断默认值
	switch dst.Type().Kind() {
	case reflect.Bool:
		boolResult, err := strconv.ParseBool(defaultName)
		if err != nil {
			dc.SetError(errors.Wrapf(err, "`default` tag value:%s error", defaultName))
			return NewDefaultProcessorValue()
		}
		return NewSingleProcessorValue(reflect.ValueOf(boolResult))
	case reflect.String:
		return NewSingleProcessorValue(reflect.ValueOf(defaultName))
	case reflect.Float64:
		floatResult, err := strconv.ParseFloat(defaultName, 64)
		if err != nil {
			dc.SetError(errors.Wrapf(err, "`default` tag value:%s error", defaultName))
			return NewDefaultProcessorValue()
		}
		return NewSingleProcessorValue(reflect.ValueOf(floatResult))
	case reflect.Int64:
		intResult, err := strconv.ParseInt(defaultName, 0, 64)
		if err != nil {
			dc.SetError(errors.Wrapf(err, "`default` tag value:%s error", defaultName))
			return NewDefaultProcessorValue()
		}
		return NewSingleProcessorValue(reflect.ValueOf(intResult))
	case reflect.Uint64:
		uintResult, err := strconv.ParseUint(defaultName, 0, 64)
		if err != nil {
			dc.SetError(errors.Wrapf(err, "`default` tag value:%s error", defaultName))
			return NewDefaultProcessorValue()
		}
		return NewSingleProcessorValue(reflect.ValueOf(uintResult))
	case reflect.Uint:
		uintResult, err := strconv.ParseUint(defaultName, 0, 64)
		if err != nil {
			dc.SetError(errors.Wrapf(err, "`default` tag value:%s error", defaultName))
			return NewDefaultProcessorValue()
		}
		return NewSingleProcessorValue(reflect.ValueOf(uint(uintResult)))
	case reflect.Int:
		intResult, err := strconv.Atoi(defaultName)
		if err != nil {
			dc.SetError(errors.Wrapf(err, "`default` tag value:%s error", defaultName))
			return NewDefaultProcessorValue()
		}
		return NewSingleProcessorValue(reflect.ValueOf(intResult))
	}

	if dc.config.StrictMode {
		dc.SetError(errors.Errorf("`default` tag %s couldn't find any valid method or field type", defaultName))
		return NewDefaultProcessorValue()
	}

	return dc.NextSelector(src, dst)
}

// method标签选择器
var MethodTagSelector = func(dc *DeepCopier, src ProcessorValue, dst ProcessorValue) ProcessorValue {
	if src.GetValueType() != ProcessorValueSingleType || dst.GetValueType() != ProcessorValueStructFieldType {
		return dc.NextSelector(src, dst)
	}

	if !src.IsValid() {
		return dc.NextSelector(src, dst)
	}

	methodName, ok := dst.GetStructFieldTag(FieldMethodTagName)
	if !ok {
		return dc.NextSelector(src, dst)
	}

	switch src.Type().Kind() {
	case reflect.Struct:
		// 指针的方法会包含结构体的方法
		ptrSrc := reflect.New(src.Type())
		ptrSrc.Elem().Set(src.GetValue())
		res := dc.CurrentSelector(NewSingleProcessorValue(ptrSrc), dst)
		if res.IsValid() {
			return res
		}
	case reflect.Ptr:
		for i := 0; i < src.NumMethod(); i++ {
			srcMethodValue := src.Method(i)

			if !srcMethodValue.IsValid() {
				continue
			}

			if src.Type().Method(i).Name != methodName {
				continue
			}

			result := srcMethodValue.Call([]reflect.Value{reflect.ValueOf(dc.args)})
			if len(result) < 1 {
				dc.SetError(errors.Errorf("`method` tag method:%s return empty", methodName))
				return NewDefaultProcessorValue()
			} else if len(result) > 1 {
				// 最后一个参数要求为error
				err, ok := result[len(result)-1].Interface().(error)
				if ok && err != nil {
					dc.SetError(err)
					return NewDefaultProcessorValue()
				}
			}

			return NewSingleProcessorValue(result[0])
		}
	}

	if dc.config.StrictMode {
		dc.SetError(errors.Errorf("couldn't find `method` tag method: %s", methodName))
		return NewDefaultProcessorValue()
	}

	return dc.NextSelector(src, dst)
}

// skip标签选择器
var SkipTagSelector = func(dc *DeepCopier, src ProcessorValue, dst ProcessorValue) ProcessorValue {
	if src.GetValueType() != ProcessorValueSingleType || dst.GetValueType() != ProcessorValueStructFieldType {
		return dc.NextSelector(src, dst)
	}

	_, ok := dst.GetStructFieldTag(SkipTagName)
	if ok {
		return NewDefaultProcessorValue()
	}

	return dc.NextSelector(src, dst)
}

// from标签选择器
var FromTagSelector = func(dc *DeepCopier, src ProcessorValue, dst ProcessorValue) ProcessorValue {
	if src.GetValueType() != ProcessorValueSingleType || dst.GetValueType() != ProcessorValueStructFieldType {
		return dc.NextSelector(src, dst)
	}

	if !src.IsValid() {
		return dc.NextSelector(src, dst)
	}

	fromName, ok := dst.GetStructFieldTag(FieldFromTagName)
	if !ok {
		return dc.NextSelector(src, dst)
	}

	switch src.Type().Kind() {
	case reflect.Ptr:
		res := dc.CurrentSelector(NewSingleProcessorValue(src.GetIndirectValue()), dst)
		if res.IsValid() {
			return res
		}
	case reflect.Struct:
		for i := 0; i < src.NumField(); i++ {
			srcFieldValue := src.Field(i)
			srcFieldType := src.Type().Field(i)

			// 匿名结构体
			if srcFieldType.Anonymous {
				res := dc.CurrentSelector(NewSingleProcessorValue(srcFieldValue), dst)
				if res.IsValid() {
					return res
				}
			}

			if srcFieldType.Name == fromName {
				return NewSingleProcessorValue(srcFieldValue)
			}
		}
	case reflect.Map:
		// key要求必须为string
		if src.Type().Key().Kind() != reflect.String {
			break
		}

		for _, i := range src.MapKeys() {
			if i.Interface().(string) == fromName && src.MapIndex(i).CanInterface() {
				// 为避免类型为interface，实际为其他类型，需二次反射
				return NewSingleProcessorValue(reflect.ValueOf(src.MapIndex(i).Interface()))
			}
		}
	}

	if dc.config.StrictMode {
		dc.SetError(errors.Errorf("couldn't find `from` tag field: %s", fromName))
		return NewDefaultProcessorValue()
	}

	return dc.NextSelector(src, dst)
}

// 默认字段名选择器
var FieldNameSelector = func(dc *DeepCopier, src ProcessorValue, dst ProcessorValue) ProcessorValue {
	if src.GetValueType() != ProcessorValueSingleType || dst.GetValueType() != ProcessorValueStructFieldType {
		return dc.NextSelector(src, dst)
	}

	if !src.IsValid() {
		return dc.NextSelector(src, dst)
	}

	dstFieldType := dst.GetStructField()
	switch src.Type().Kind() {
	case reflect.Ptr:
		res := dc.CurrentSelector(NewSingleProcessorValue(src.GetIndirectValue()), dst)
		if res.IsValid() {
			return res
		}
	case reflect.Struct:
		for i := 0; i < src.NumField(); i++ {
			srcFieldValue := src.Field(i)
			srcFieldType := src.Type().Field(i)

			// 匿名结构体
			if srcFieldType.Anonymous {
				res := dc.CurrentSelector(NewSingleProcessorValue(srcFieldValue), dst)
				if res.IsValid() {
					return res
				}
			}

			if srcFieldType.Name == dstFieldType.Name {
				return NewSingleProcessorValue(srcFieldValue)
			}
		}
	case reflect.Map:
		// key要求必须为string
		if src.Type().Key().Kind() != reflect.String {
			break
		}

		for _, i := range src.MapKeys() {
			if i.Interface().(string) == dstFieldType.Name && src.MapIndex(i).CanInterface() {
				// 为避免类型为interface，实际为其他类型，需二次反射
				return NewSingleProcessorValue(reflect.ValueOf(src.MapIndex(i).Interface()))
			}
		}
	}

	return dc.NextSelector(src, dst)
}
