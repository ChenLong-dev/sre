package deepcopy

import (
	"database/sql/driver"
	"fmt"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"reflect"
	"time"
)

// 转换器方法
type TransformerFunc func(dc *DeepCopier, candidate, dst ProcessorValue) ProcessorValue

// timeformat标签转换器
var TimeFormatTagTransformer = func(dc *DeepCopier, candidate, dst ProcessorValue) ProcessorValue {
	if candidate.GetValueType() != ProcessorValueSingleType || dst.GetValueType() != ProcessorValueStructFieldType {
		return dc.NextTransformer(candidate, dst)
	}

	srcFormat, ok := dst.GetStructFieldTag(TimeFormatTagName)
	if !ok {
		return dc.NextTransformer(candidate, dst)
	}

	if !candidate.IsValid() || !candidate.CanInterface() {
		return dc.NextTransformer(candidate, dst)
	}

	// 为避免类型为interface，实际为时间ptr类型，需二次反射
	indirectValue := reflect.Indirect(reflect.ValueOf(candidate.Interface()))
	if !indirectValue.IsValid() {
		return dc.NextTransformer(candidate, dst)
	}

	t, ok := indirectValue.Interface().(time.Time)
	if !ok {
		if dc.config.StrictMode {
			dc.SetError(errors.Errorf("`timeformat` tag field: %s isn't time.Time", candidate.Type().Name()))
			return NewDefaultProcessorValue()
		}

		return dc.NextTransformer(candidate, dst)
	}

	return dc.NextTransformer(
		NewSingleProcessorValue(reflect.ValueOf(t.In(time.Local).Format(srcFormat))),
		dst,
	)
}

// primitive.ObjectID/string转换器
var ObjectIDFormatTagTransformer = func(dc *DeepCopier, candidate, dst ProcessorValue) ProcessorValue {
	if candidate.GetValueType() != ProcessorValueSingleType || dst.GetValueType() != ProcessorValueStructFieldType {
		return dc.NextTransformer(candidate, dst)
	}

	_, ok := dst.GetStructFieldTag(ObjectIDTagName)
	if !ok {
		return dc.NextTransformer(candidate, dst)
	}

	if !candidate.IsValid() || !candidate.CanInterface() {
		return dc.NextTransformer(candidate, dst)
	}

	indirectValue := reflect.Indirect(reflect.ValueOf(candidate.Interface()))
	if !indirectValue.IsValid() {
		return dc.NextTransformer(candidate, dst)
	}

	if candidate.Type().Kind() == reflect.String {
		indirectValue := reflect.Indirect(reflect.ValueOf(candidate.Interface()))
		if !indirectValue.IsValid() {
			return dc.NextTransformer(candidate, dst)
		}
		s, ok := indirectValue.Interface().(string)
		if !ok {
			if dc.config.StrictMode {
				dc.SetError(errors.Errorf("`objectid` tag field: %v isn't string.", indirectValue.Interface()))
				return NewDefaultProcessorValue()
			}
			return dc.NextTransformer(candidate, dst)
		}
		objectID, err := primitive.ObjectIDFromHex(s)
		if err != nil {
			if dc.config.StrictMode {
				dc.SetError(errors.Errorf("`objectid` tag field: %s is not a valid ObjectID.", s))
				return NewDefaultProcessorValue()
			}
			return dc.NextTransformer(candidate, dst)
		}

		return dc.NextTransformer(
			NewSingleProcessorValue(reflect.ValueOf(objectID)),
			dst,
		)
	} else if dst.Type().Kind() == reflect.String {
		indirectValue := reflect.Indirect(reflect.ValueOf(candidate.Interface()))
		if !indirectValue.IsValid() {
			return dc.NextTransformer(candidate, dst)
		}
		objectID, ok := indirectValue.Interface().(primitive.ObjectID)
		if !ok {
			if dc.config.StrictMode {
				dc.SetError(errors.Errorf("`objectid` tag field: %v isn't primitive.ObjectID.", indirectValue.Interface()))
				return NewDefaultProcessorValue()
			}
			return dc.NextTransformer(candidate, dst)
		}

		return dc.NextTransformer(
			NewSingleProcessorValue(reflect.ValueOf(objectID.Hex())),
			dst,
		)
	} else {
		return dc.NextTransformer(candidate, dst)
	}
}

// null包/driver.Valuer转换器
var SqlTransformer = func(dc *DeepCopier, candidate, dst ProcessorValue) ProcessorValue {
	// 避免指针类型
	candidateIndirect := candidate.GetIndirectValue()
	dstIndirect := dst.GetIndirectValue()
	if !candidateIndirect.IsValid() || !dstIndirect.IsValid() {
		return dc.NextTransformer(candidate, dst)
	}

	// 候选值为null包
	if candidateIndirect.Type().ConvertibleTo(reflect.TypeOf((*driver.Valuer)(nil)).Elem()) {
		valuerValue, _ := candidate.Interface().(driver.Valuer).Value()
		if valuerValue == nil {
			return dc.NextTransformer(candidate, dst)
		}
		res := reflect.ValueOf(valuerValue)

		// int类型特殊处理
		if dst.Type().Kind() == reflect.Int && res.Type().Kind() == reflect.Int64 {
			return dc.NextTransformer(NewSingleProcessorValue(reflect.ValueOf(int(valuerValue.(int64)))), dst)
		}

		return dc.NextTransformer(NewSingleProcessorValue(res), dst)
	} else if dstIndirect.Type().ConvertibleTo(reflect.TypeOf((*driver.Valuer)(nil)).Elem()) { // 目标值为null包
		ptr := reflect.New(dstIndirect.Type())

		// 避免类型不一致
		valueMethod := reflect.Indirect(ptr).MethodByName("ValueOrZero")
		if !valueMethod.IsValid() {
			return dc.NextTransformer(candidate, dst)
		}
		valueMethodRes := valueMethod.Call([]reflect.Value{})
		if len(valueMethodRes) != 1 || !candidate.Type().ConvertibleTo(valueMethodRes[0].Type()) {
			return dc.NextTransformer(candidate, dst)
		}

		setMethod := ptr.MethodByName("SetValid")
		if !setMethod.IsValid() {
			return dc.NextTransformer(candidate, dst)
		}
		setMethod.Call([]reflect.Value{candidate.GetValue()})

		return dc.NextTransformer(NewSingleProcessorValue(ptr), dst)
	} else {
		return dc.NextTransformer(candidate, dst)
	}
}

// string标签转换器
var StringTagTransformer = func(dc *DeepCopier, candidate, dst ProcessorValue) ProcessorValue {
	if candidate.GetValueType() != ProcessorValueSingleType || dst.GetValueType() != ProcessorValueStructFieldType {
		return dc.NextTransformer(candidate, dst)
	}

	_, ok := dst.GetStructFieldTag(StringTagName)
	if !ok {
		return dc.NextTransformer(candidate, dst)
	}

	if !candidate.IsValid() || !candidate.CanInterface() {
		return dc.NextTransformer(candidate, dst)
	}

	if dst.Type().Kind() != reflect.String {
		if dc.config.StrictMode {
			dc.SetError(errors.Errorf("`string` tag field is %s, isn't string", dst.Type()))
			return NewDefaultProcessorValue()
		} else {
			return dc.NextTransformer(candidate, dst)
		}
	}

	return dc.NextTransformer(
		NewSingleProcessorValue(reflect.ValueOf(fmt.Sprintf("%v", candidate.Interface()))),
		dst,
	)
}

// bool标签转换器
var BoolTagTransformer = func(dc *DeepCopier, candidate, dst ProcessorValue) ProcessorValue {
	if candidate.GetValueType() != ProcessorValueSingleType || dst.GetValueType() != ProcessorValueStructFieldType {
		return dc.NextTransformer(candidate, dst)
	}

	_, ok := dst.GetStructFieldTag(BoolTagName)
	if !ok {
		return dc.NextTransformer(candidate, dst)
	}

	if !candidate.IsValid() || !candidate.CanInterface() {
		return dc.NextTransformer(candidate, dst)
	}

	if dst.Type().Kind() != reflect.Bool {
		if dc.config.StrictMode {
			dc.SetError(errors.Errorf("`bool` tag field is %s, isn't bool", dst.Type()))
			return NewDefaultProcessorValue()
		} else {
			return dc.NextTransformer(candidate, dst)
		}
	}

	res := false
	switch candidate.Type().Kind() {
	case reflect.Float64:
		data, ok := candidate.Interface().(float64)
		if !ok {
			break
		}
		res = data > 0
	case reflect.Int:
		data, ok := candidate.Interface().(int)
		if !ok {
			break
		}
		res = data > 0
	case reflect.Int64:
		data, ok := candidate.Interface().(int64)
		if !ok {
			break
		}
		res = data > 0
	case reflect.Uint64:
		data, ok := candidate.Interface().(uint64)
		if !ok {
			break
		}
		res = data > 0
	case reflect.Uint:
		data, ok := candidate.Interface().(uint)
		if !ok {
			break
		}
		res = data > 0
	}

	return dc.NextTransformer(
		NewSingleProcessorValue(reflect.ValueOf(res)),
		dst,
	)
}
