package deepcopy

import (
	"database/sql/driver"
	"fmt"
	"github.com/pkg/errors"
	"reflect"
	"strings"
)

// 深拷贝
// 用于将源结构体拷贝至目标结构体
type DeepCopier struct {
	// 配置文件
	config *Config
	// 目标结构体
	dst interface{}
	// 源结构体
	src interface{}
	// 额外参数
	args map[string]interface{}

	// 选择器链
	selectorChain []SelectorFunc
	// 选择器处理索引
	selectorIndex int
	// 转换器链
	transformerChain []TransformerFunc
	// 转换器处理索引
	transformerIndex int

	// 当前错误
	err error
}

// 从源结构体初始化
func Copy(src interface{}) *DeepCopier {
	return &DeepCopier{
		src:    src,
		config: &Config{},
	}
}

// 设置配置文件
func (dc *DeepCopier) SetConfig(config *Config) *DeepCopier {
	if config != nil {
		dc.config = config
	}
	return dc
}

// 增加额外参数
func (dc *DeepCopier) AddArg(key string, value interface{}) *DeepCopier {
	if dc.args == nil {
		dc.args = make(map[string]interface{})
	}
	dc.args[key] = value
	return dc
}

// 配置额外参数
func (dc *DeepCopier) WithArgs(args map[string]interface{}) *DeepCopier {
	dc.args = args
	return dc
}

// 拷贝至目标结构体
func (dc *DeepCopier) To(dst interface{}) error {
	dc.dst = dst
	return dc.process()
}

// 从源结构体拷贝，此前copy函数内的结构体为目标结构体
func (dc *DeepCopier) From(src interface{}) error {
	dc.dst = dc.src
	dc.src = src
	return dc.process()
}

// 处理
func (dc *DeepCopier) process() error {
	dstValue := reflect.Indirect(reflect.ValueOf(dc.dst))
	if !dstValue.CanAddr() {
		return errors.New(fmt.Sprintf("dst %+v is unaddressable", dstValue.Interface()))
	}

	dc.selectorIndex = -1
	dc.selectorChain = []SelectorFunc{
		SkipTagSelector,
		MethodTagSelector,
		FromTagSelector,
		ToTagSelector,
		FieldNameSelector,
		DefaultTagSelector,
	}
	dc.transformerIndex = -1
	dc.transformerChain = []TransformerFunc{
		SqlTransformer,
		TimeFormatTagTransformer,
		StringTagTransformer,
		BoolTagTransformer,
		ObjectIDFormatTagTransformer,
	}

	dc.copy(reflect.ValueOf(dc.src), reflect.ValueOf(dc.dst))
	return dc.err
}

// 处理当前转换器
func (dc *DeepCopier) CurrentTransformer(candidate, dst ProcessorValue) ProcessorValue {
	curIndex := dc.transformerIndex
	defer func() {
		dc.transformerIndex = curIndex
	}()

	if dc.transformerIndex < len(dc.transformerChain) {
		return dc.transformerChain[dc.transformerIndex](dc, candidate, dst)
	}

	return candidate
}

// 处理当前选择器
func (dc *DeepCopier) CurrentSelector(src ProcessorValue, dst ProcessorValue) ProcessorValue {
	curIndex := dc.selectorIndex
	defer func() {
		dc.selectorIndex = curIndex
	}()

	if dc.selectorIndex < len(dc.selectorChain) {
		return dc.selectorChain[dc.selectorIndex](dc, src, dst)
	}

	return NewDefaultProcessorValue()
}

// 处理下一个转换器
func (dc *DeepCopier) NextTransformer(candidate, dst ProcessorValue) ProcessorValue {
	dc.transformerIndex++
	if dc.transformerIndex < len(dc.transformerChain) {
		return dc.transformerChain[dc.transformerIndex](dc, candidate, dst)
	}

	return candidate
}

// 处理下一个选择器
func (dc *DeepCopier) NextSelector(src ProcessorValue, dst ProcessorValue) ProcessorValue {
	dc.selectorIndex++
	if dc.selectorIndex < len(dc.selectorChain) {
		return dc.selectorChain[dc.selectorIndex](dc, src, dst)
	}

	return NewDefaultProcessorValue()
}

// 选择并转换源数据
func (dc *DeepCopier) SelectAndTransformSourceValue(src, dst ProcessorValue) ProcessorValue {
	dc.selectorIndex = -1
	candidate := dc.NextSelector(src, dst)
	if !candidate.IsValid() {
		return NewDefaultProcessorValue()
	}

	res := dc.TransformSourceValue(candidate, dst)

	return res
}

// 转换原数据
func (dc *DeepCopier) TransformSourceValue(candidate, dst ProcessorValue) ProcessorValue {
	dc.transformerIndex = -1
	res := dc.NextTransformer(candidate, dst)
	if !res.IsValid() {
		return NewDefaultProcessorValue()
	}
	return res
}

// 是否启用指定可选标签
func (dc *DeepCopier) IsEnableOptional(tagName string) bool {
	for _, name := range dc.config.EnableOptionalTags {
		if name == tagName {
			return true
		}
	}
	return false
}

// 设置错误
func (dc *DeepCopier) SetError(err error) {
	dc.err = err
}

// 获取目标的key名
// 要求目标为Map，并且源为结构体
func (dc *DeepCopier) getDstMapName(src reflect.Value, index int) string {
	srcFieldType := src.Type().Field(index)

	keyName := srcFieldType.Name
	if !dc.IsEnableOptional(MapKeyTagName) {
		return keyName
	}

	// 降低取tag频率
	if !strings.Contains(srcFieldType.Tag.Get(TagName), MapKeyTagName) {
		return keyName
	}

	// todo:频繁取tag，会导致cpu过高
	tag := getTags(srcFieldType)
	defer func() {
		cleanTagsPool(tag)
	}()
	tagKey, ok := tag[MapKeyTagName]
	if !ok {
		return keyName
	}

	return tagKey
}

// 赋值前校验
func (dc *DeepCopier) beforeSetCheck(src reflect.Value) bool {
	if !src.IsValid() {
		return false
	}

	// 检查非零模式
	if dc.config.NotZeroMode {
		if src.IsZero() {
			return false
		}

		// 针对null类型
		srcIndirect := reflect.Indirect(src)
		if srcIndirect.Type().ConvertibleTo(reflect.TypeOf((*driver.Valuer)(nil)).Elem()) {
			isZero := srcIndirect.MethodByName("IsZero")
			if !isZero.IsValid() {
				return true
			}
			isZeroRes := isZero.Call([]reflect.Value{})
			if len(isZeroRes) == 1 {
				return !isZeroRes[0].Interface().(bool)
			}
		}
	}

	return true
}

// 直接赋值
// isTransform 参数用来判断是否需要转换器
func (dc *DeepCopier) directAssign(src reflect.Value, dst reflect.Value, needTransform bool) bool {
	if !dc.config.FullTraversalMode || reflect.Indirect(src).Kind() != reflect.Struct {
		// 源值可直接赋值
		if src.IsValid() && dst.CanSet() && src.Type().AssignableTo(dst.Type()) {
			dst.Set(src)
			return true
		}

		// 源指针的值可直接赋值
		indirectValue := reflect.Indirect(src)
		if indirectValue.IsValid() && dst.CanSet() && indirectValue.Type().AssignableTo(dst.Type()) {
			dst.Set(indirectValue)
			return true
		}
	}

	// 转换器
	if needTransform && src.IsValid() {
		tranRes := dc.TransformSourceValue(NewSingleProcessorValue(src), NewSingleProcessorValue(dst))
		return dc.directAssign(tranRes.GetValue(), dst, false)
	}

	return false
}

// 拷贝
func (dc *DeepCopier) copy(src reflect.Value, dst reflect.Value) {
	if dc.err != nil || !dst.IsValid() {
		return
	}

	// 预先校验，若不通过则直接返回
	if !dc.beforeSetCheck(src) {
		return
	}

	// 尝试直接赋值，如果成功，则返回
	isAssigned := dc.directAssign(src, dst, true)
	if isAssigned {
		return
	}

	// 源类型为指针时，取值并递归
	if src.Kind() == reflect.Ptr {
		dc.copy(reflect.Indirect(src), dst)
		return
	}

	// 判断类型并递归
	switch dst.Kind() {
	case reflect.Slice, reflect.Array:
		// 要求源值为数组/切片
		if src.Kind() != reflect.Slice && src.Kind() != reflect.Array {
			return
		}

		res := reflect.MakeSlice(dst.Type(), 0, 0)
		for i := 0; i < src.Len(); i++ {
			cur := reflect.New(dst.Type().Elem())
			dc.copy(src.Index(i), cur)
			res = reflect.Append(res, reflect.Indirect(cur))
		}

		if dst.CanSet() && res.Type().AssignableTo(dst.Type()) && res.IsValid() {
			dst.Set(res)
		}
	case reflect.Map:
		// 要求源值为map/结构体
		if src.Kind() != reflect.Map && src.Kind() != reflect.Struct {
			return
		}

		if src.Kind() == reflect.Map {
			// key类型要求一致
			if src.Type().Key() != dst.Type().Key() {
				return
			}

			res := reflect.MakeMap(dst.Type())
			for _, i := range src.MapKeys() {
				value := reflect.New(dst.Type().Elem())
				dc.copy(src.MapIndex(i), value)
				res.SetMapIndex(i, reflect.Indirect(value))
			}

			if dst.CanSet() && res.Type().AssignableTo(dst.Type()) && res.IsValid() {
				dst.Set(res)
			}
		} else if src.Kind() == reflect.Struct {
			// key要求为字符串
			if dst.Type().Key().Kind() != reflect.String {
				return
			}

			res := reflect.MakeMap(dst.Type())
			for i := 0; i < src.NumField(); i++ {
				value := reflect.New(dst.Type().Elem())
				srcField := src.Field(i)
				// 预先校验结构体字段
				if !dc.beforeSetCheck(srcField) {
					continue
				}

				dc.copy(srcField, value)

				// 获取键名
				keyName := dc.getDstMapName(src, i)
				res.SetMapIndex(reflect.ValueOf(keyName), reflect.Indirect(value))
			}

			if dst.CanSet() && res.Type().AssignableTo(dst.Type()) && res.IsValid() {
				dst.Set(res)
			}
		}
	case reflect.Struct:
		for i := 0; i < dst.NumField(); i++ {
			dstFieldValue := dst.Field(i)
			if !dstFieldValue.CanSet() {
				continue
			}

			dstField := NewStructFieldProcessorValue(dst, i)
			// 选择并转换源值
			cur := dc.SelectAndTransformSourceValue(NewSingleProcessorValue(src), dstField)
			// 使用后清理工作
			dstField.CleanUp()

			if !cur.IsValid() {
				// 匿名结构体
				if dst.Type().Field(i).Anonymous {
					dc.copy(src, dstFieldValue)
				}
			} else {
				dc.copy(cur.GetValue(), dstFieldValue)
			}
		}
	case reflect.Ptr:
		if dst.IsNil() {
			if dst.CanSet() {
				ptr := reflect.New(dst.Type().Elem())
				dc.copy(src, reflect.Indirect(ptr))
				dst.Set(ptr)
			}
		} else {
			dc.copy(src, reflect.Indirect(dst))
		}
	}
}
