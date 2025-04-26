package deepcopy

import "reflect"

// 处理器值类型
type ProcessorValueType string

const (
	// 简单类型
	ProcessorValueSingleType ProcessorValueType = "single"
	// 结构体字段类型
	ProcessorValueStructFieldType ProcessorValueType = "struct-field"
)

// 处理器值
type ProcessorValue struct {
	// 值
	reflect.Value

	// 类型
	t ProcessorValueType

	// 结构体值
	parentStruct reflect.Value
	// 结构体索引
	fieldIndex int
	// 结构体字段标签
	fieldTags Tags
}

// 新建默认类型
func NewDefaultProcessorValue() ProcessorValue {
	return ProcessorValue{
		t:     ProcessorValueSingleType,
		Value: reflect.Value{},
	}
}

// 新建简单类型
func NewSingleProcessorValue(v reflect.Value) ProcessorValue {
	return ProcessorValue{
		t:     ProcessorValueSingleType,
		Value: v,
	}
}

// 新建结构体类型
func NewStructFieldProcessorValue(parentStruct reflect.Value, index int) ProcessorValue {
	tags := getTags(parentStruct.Type().Field(index))
	return ProcessorValue{
		t:            ProcessorValueStructFieldType,
		Value:        parentStruct.Field(index),
		parentStruct: parentStruct,
		fieldIndex:   index,
		fieldTags:    tags,
	}
}

// 清理
func (v *ProcessorValue) CleanUp() {
	if v.fieldTags != nil {
		cleanTagsPool(v.fieldTags)
	}
	return
}

func (v *ProcessorValue) GetValueType() ProcessorValueType {
	return v.t
}

func (v *ProcessorValue) GetValue() reflect.Value {
	return v.Value
}

func (v *ProcessorValue) GetIndirectValue() reflect.Value {
	return reflect.Indirect(v.GetValue())
}

func (v *ProcessorValue) GetParentStruct() reflect.Value {
	if v.GetValueType() == ProcessorValueStructFieldType {
		return v.parentStruct
	}

	return reflect.Value{}
}

func (v *ProcessorValue) GetStructField() reflect.StructField {
	if v.t == ProcessorValueStructFieldType {
		return v.GetParentStruct().Type().Field(v.fieldIndex)
	}

	return reflect.StructField{}
}

func (v *ProcessorValue) GetStructFieldTag(tagName string) (string, bool) {
	if v.t == ProcessorValueStructFieldType {
		tagValue, ok := v.fieldTags[tagName]
		return tagValue, ok
	}

	return "", false
}
