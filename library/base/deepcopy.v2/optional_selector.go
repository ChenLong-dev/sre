package deepcopy

import (
	"reflect"
	"strings"
)

var ToTagSelector = func(dc *DeepCopier, src ProcessorValue, dst ProcessorValue) ProcessorValue {
	if src.GetValueType() != ProcessorValueSingleType || dst.GetValueType() != ProcessorValueStructFieldType {
		return dc.NextSelector(src, dst)
	}

	if !src.IsValid() {
		return dc.NextSelector(src, dst)
	}

	if !dc.IsEnableOptional(FieldToTagName) {
		return dc.NextSelector(src, dst)
	}

	dstTypeField := dst.GetStructField()
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

			// 降低取tag频率
			if !strings.Contains(srcFieldType.Tag.Get(TagName), FieldToTagName) {
				continue
			}
			// todo:频繁取tag，会导致cpu过高
			tag := getTags(srcFieldType)
			toName, ok := tag[FieldToTagName]
			cleanTagsPool(tag)
			if !ok {
				continue
			}

			if srcFieldType.Anonymous {
				res := dc.CurrentSelector(NewSingleProcessorValue(srcFieldValue), dst)
				if res.IsValid() {
					return res
				}
			}

			if toName == dstTypeField.Name {
				return NewSingleProcessorValue(srcFieldValue)
			}
		}
	}

	return dc.NextSelector(src, dst)
}
