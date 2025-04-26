package deepcopy

import (
	"database/sql/driver"
	"fmt"
	"github.com/pkg/errors"
	"reflect"
	"strconv"
	"sync"
	"time"
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
}

// 从源结构体初始化
// Deprecated
// 推荐使用v2版本
func Copy(src interface{}) *DeepCopier {
	return &DeepCopier{
		src: src,
		config: &Config{
			ParseAnonymousStruct: false,
		},
	}
}

// 设置配置文件
// Deprecated
// 推荐使用v2版本
func (dc *DeepCopier) SetConfig(config *Config) *DeepCopier {
	dc.config = config
	return dc
}

// 增加额外参数
// Deprecated
// 推荐使用v2版本
func (dc *DeepCopier) AddArg(key string, value interface{}) *DeepCopier {
	if dc.args == nil {
		dc.args = make(map[string]interface{})
	}
	dc.args[key] = value
	return dc
}

// 配置额外参数
// Deprecated
// 推荐使用v2版本
func (dc *DeepCopier) WithArgs(args map[string]interface{}) *DeepCopier {
	dc.args = args
	return dc
}

// 拷贝至目标结构体
// Deprecated
// 推荐使用v2版本
func (dc *DeepCopier) To(dst interface{}) error {
	dc.dst = dst
	return dc.process()
}

// 从源结构体拷贝，此前copy函数内的结构体为目标结构体
// Deprecated
// 推荐使用v2版本
func (dc *DeepCopier) From(src interface{}) error {
	dc.dst = dc.src
	dc.src = src
	return dc.process()
}

// 处理
func (dc *DeepCopier) process() error {
	var (
		dstValue = reflect.Indirect(reflect.ValueOf(dc.dst))
	)

	if !dstValue.CanAddr() {
		return errors.New(fmt.Sprintf("dst %+v is unaddressable", dstValue.Interface()))
	}

	err := dc.processField()
	if err != nil {
		return err
	}

	return nil
}

// 处理字段
func (dc *DeepCopier) processField() error {
	var (
		dstValue      = reflect.Indirect(reflect.ValueOf(dc.dst))
		dstFieldNames = dc.getFieldNames(fieldNamePool.Get().([]string), dc.dst)
		dstTagsMap    = getTagsMap(dstValue, dstFieldNames)

		srcValue      = reflect.Indirect(reflect.ValueOf(dc.src))
		srcFieldNames = dc.getFieldNames(fieldNamePool.Get().([]string), dc.src)
		srcTagsMap    = getTagsMap(srcValue, srcFieldNames)

		fieldToTagMap = getFieldToTagMap(srcTagsMap)
	)

	if dstValue.Type().Kind() != reflect.Struct {
		return errors.New("dst can't get struct")
	}
	if srcValue.Type().Kind() != reflect.Struct {
		return errors.New("src can't get struct")
	}

	srcMethodResultMap, err := dc.getSrcMethodResultMap(dstTagsMap)
	if err != nil {
		return err
	}
	dstDefaultMethodResultMap, err := dc.getDstDefaultMethodResultMap(dstTagsMap)
	if err != nil {
		return err
	}

	for _, fieldName := range dstFieldNames {
		var (
			dstFieldValue               = dstValue.FieldByName(fieldName)
			dstFieldType, dstFieldFound = dstValue.Type().FieldByName(fieldName)
			dstFieldName                = dstFieldType.Name
			dstTags                     = dstTagsMap[dstFieldName]

			srcFieldName = dstFieldName
		)
		if !dstFieldFound {
			continue
		}

		// skip 标签判断
		isFinish, err := dc.processSkipTag(dstTags)
		if err != nil {
			return err
		} else if isFinish {
			continue
		}

		// field 标签判断，标签优先级：method>from>to
		isFinish, err = dc.processMethodTag(srcMethodResultMap, dstFieldValue, dstFieldName)
		if err != nil {
			return err
		} else if isFinish {
			continue
		} else if fromName, ok := dstTags[FieldFromTagName]; ok && fromName != "" {
			srcFieldName = fromName
		} else if srcToName, ok := fieldToTagMap[dstFieldName]; ok && srcToName != "" {
			srcFieldName = srcToName
		}

		var (
			_, srcFieldFound = srcValue.Type().FieldByName(srcFieldName)
			srcFieldValue    = srcValue.FieldByName(srcFieldName)
		)
		if !srcFieldFound {
			// default 标签判断
			isFinish, err = dc.processDefaultMethodTag(dstDefaultMethodResultMap, dstFieldValue, dstFieldName)
			if err != nil {
				return err
			}

			continue
		}

		// force 标签判断
		isFinish, err = dc.processForceTag(dstTags, srcFieldValue, dstFieldValue)
		if err != nil {
			return err
		} else if isFinish {
			continue
		}

		// sql 处理
		isFinish, err = dc.processSqlField(dstTags, srcFieldValue, dstFieldValue)
		if err != nil {
			return err
		} else if isFinish {
			continue
		}

		// time 标签判断
		isFinish, err = dc.processTimeFormatTag(dstTags, srcFieldValue, dstFieldValue)
		if err != nil {
			return err
		} else if isFinish {
			continue
		}

		// others
		_, err = dc.processOthers(dstTags, srcFieldValue, dstFieldValue)
		if err != nil {
			return err
		}
	}

	cleanKVMap(fieldToTagMap)
	cleanReflectValueMap(srcMethodResultMap)
	cleanReflectValueMap(dstDefaultMethodResultMap)
	cleanFieldTags(dstTagsMap)
	cleanFieldTags(srcTagsMap)
	cleanFieldNameSlice(dstFieldNames)
	cleanFieldNameSlice(srcFieldNames)

	return nil
}

// 反射获取结构体内字段对
// key为字段名，value为反射值
func (dc *DeepCopier) getFieldNames(fields []string, instance interface{}) []string {
	var (
		v = reflect.Indirect(reflect.ValueOf(instance))
		t = v.Type()
	)

	if t.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < v.NumField(); i++ {
		var (
			vField = v.Field(i)
			tField = v.Type().Field(i)
		)

		// 若不可导出，则不复制
		if tField.PkgPath != "" {
			continue
		}

		// 解析出匿名结构体中字段
		if dc.config.ParseAnonymousStruct && tField.Type.Kind() == reflect.Struct && tField.Anonymous {
			childInterface := vField.Addr().Interface()
			fields = dc.getFieldNames(fields, childInterface)
			continue
		}

		fields = append(fields, tField.Name)
	}

	return fields
}

func (dc *DeepCopier) processMethodTag(srcMethodResultMap map[string][]reflect.Value, dstFieldValue reflect.Value, dstFieldName string) (bool, error) {
	result, ok := srcMethodResultMap[dstFieldName]
	if !ok {
		return false, nil
	}

	if len(result) > 1 {
		err, ok := result[len(result)-1].Interface().(error)
		if ok && err != nil {
			return false, err.(error)
		}
	}

	var (
		resultInterface = result[0].Interface()
		resultValue     = reflect.ValueOf(resultInterface)
		resultType      = resultValue.Type()
	)

	// Ptr -> Value/Ptr
	if resultValue.Kind() == reflect.Ptr {
		if resultValue.Elem().Type().AssignableTo(dstFieldValue.Type()) {
			dstFieldValue.Set(resultValue.Elem())
			return true, nil
		}
	}

	// Value -> Ptr
	if dstFieldValue.Kind() == reflect.Ptr {
		ptr := reflect.New(resultType)
		ptr.Elem().Set(resultValue)

		if ptr.Type().AssignableTo(dstFieldValue.Type()) {
			dstFieldValue.Set(ptr)
			return true, nil
		}
	}

	// Others
	if resultType.AssignableTo(dstFieldValue.Type()) && result[0].IsValid() {
		dstFieldValue.Set(result[0])
		return true, nil
	}

	return false, nil
}
func (dc *DeepCopier) processDefaultMethodTag(dstDefaultMethodResultMap map[string][]reflect.Value, dstFieldValue reflect.Value, dstFieldName string) (bool, error) {
	result, ok := dstDefaultMethodResultMap[dstFieldName]
	if !ok {
		return false, nil
	}

	if len(result) > 1 {
		err, ok := result[len(result)-1].Interface().(error)
		if ok && err != nil {
			return false, err.(error)
		}
	}

	var (
		resultInterface = result[0].Interface()
		resultValue     = reflect.ValueOf(resultInterface)
		resultType      = resultValue.Type()
	)

	// Ptr -> Value/Ptr
	if resultValue.Kind() == reflect.Ptr {
		if resultValue.Elem().Type().AssignableTo(dstFieldValue.Type()) {
			dstFieldValue.Set(resultValue.Elem())
			return true, nil
		}
	}

	// Value -> Ptr
	if dstFieldValue.Kind() == reflect.Ptr {
		ptr := reflect.New(resultType)
		ptr.Elem().Set(resultValue)

		if ptr.Type().AssignableTo(dstFieldValue.Type()) {
			dstFieldValue.Set(ptr)
			return true, nil
		}
	}

	// Others且可赋值
	if resultType.AssignableTo(dstFieldValue.Type()) && resultValue.IsValid() {
		dstFieldValue.Set(result[0])
		return true, nil
	}

	// 标签值的方法不存在，需手动判断，并转换
	realResult, ok := resultInterface.(string)
	if !ok {
		return false, errors.New(fmt.Sprintf("method result %s is not string", resultInterface))
	}

	switch dstFieldValue.Type().Kind() {
	case reflect.Bool:
		boolResult, err := strconv.ParseBool(realResult)
		if err != nil {
			return false, err
		}
		resultInterface = boolResult
	case reflect.String:
		resultInterface = realResult
	case reflect.Float64:
		floatResult, err := strconv.ParseFloat(realResult, 64)
		if err != nil {
			return false, err
		}
		resultInterface = floatResult
	case reflect.Int:
		intResult, err := strconv.Atoi(realResult)
		if err != nil {
			return false, err
		}
		resultInterface = intResult
	default:
		return false, nil
	}

	resultValue = reflect.ValueOf(resultInterface)
	if resultValue.Type().AssignableTo(dstFieldValue.Type()) && resultValue.IsValid() {
		dstFieldValue.Set(resultValue)
		return true, nil
	}

	return false, nil
}

// 获取src源函数中与dst目标method标签名相同的函数执行结果映射表，key为dst字段名，value为函数执行结果
func (dc *DeepCopier) getSrcMethodResultMap(dstTagsMap map[string]Tags) (map[string][]reflect.Value, error) {
	var (
		srcMethodNames     = getMethodNames(dc.src)
		srcMethodResultMap = reflectValueMapPool.Get().(map[string][]reflect.Value)

		dstMethodMap = kvMapPool.Get().(map[string]string)
	)

	// 获取目标method标签的map
	for dstFieldName, tags := range dstTagsMap {
		if methodName, ok := tags[FieldMethodTagName]; ok && methodName != "" {
			dstMethodMap[methodName] = dstFieldName
		}
	}

	for _, m := range srcMethodNames {
		dstField, ok := dstMethodMap[m]
		if !ok {
			continue
		}

		method := reflect.ValueOf(dc.src).MethodByName(m)
		if !method.IsValid() {
			return nil, errors.New(fmt.Sprintf("method %s is invalid", m))
		}

		// 调用函数
		args := []reflect.Value{reflect.ValueOf(dc.args)}
		result := method.Call(args)
		if len(result) < 1 {
			return nil, errors.New(fmt.Sprintf("method %s return empty", m))
		}

		srcMethodResultMap[dstField] = result
	}

	cleanKVMap(dstMethodMap)

	return srcMethodResultMap, nil
}

// 获取dst标签中存在默认函数标签的函数执行结果映射表，key为dst字段名，value为函数执行结果
func (dc *DeepCopier) getDstDefaultMethodResultMap(dstTagsMap map[string]Tags) (map[string][]reflect.Value, error) {
	var (
		dstMethodNames     = getMethodNames(dc.dst)
		dstMethodResultMap = reflectValueMapPool.Get().(map[string][]reflect.Value)

		dstMethodMap = kvMapPool.Get().(map[string]string)
	)

	// 获取目标method标签的map
	for dstFieldName, tags := range dstTagsMap {
		if methodName, ok := tags[DefaultMethodName]; ok && methodName != "" {
			dstMethodMap[methodName] = dstFieldName
			dstMethodResultMap[dstFieldName] = []reflect.Value{reflect.ValueOf(methodName)}
		}
	}

	for _, m := range dstMethodNames {
		dstField, ok := dstMethodMap[m]
		if !ok {
			continue
		}

		method := reflect.ValueOf(dc.dst).MethodByName(m)
		if !method.IsValid() {
			return nil, errors.New(fmt.Sprintf("method %s is invalid", m))
		}

		// 调用函数
		args := []reflect.Value{reflect.ValueOf(dc.args)}
		result := method.Call(args)
		if len(result) < 1 {
			return nil, errors.New(fmt.Sprintf("method %s return empty", m))
		}

		dstMethodResultMap[dstField] = result
	}

	cleanKVMap(dstMethodMap)

	return dstMethodResultMap, nil
}

func (dc *DeepCopier) processSkipTag(tags Tags) (bool, error) {
	if _, ok := tags[SkipTagName]; ok {
		return true, nil
	} else {
		return false, nil
	}
}

func (dc *DeepCopier) processForceTag(tags Tags, srcFieldValue, dstFieldValue reflect.Value) (bool, error) {
	if _, ok := tags[ForceTagName]; !ok {
		return false, nil
	}

	if dstFieldValue.Kind() == reflect.Interface {
		dstFieldValue.Set(srcFieldValue)
		return true, nil
	}

	return false, nil
}

func (dc *DeepCopier) processTimeFormatTag(tags Tags, srcFieldValue, dstFieldValue reflect.Value) (bool, error) {
	formatString, ok := tags[TimeFormatTagName]
	if !ok || formatString == "" {
		return false, nil
	}

	srcIndirectValue := reflect.Indirect(srcFieldValue)
	if !srcIndirectValue.IsValid() {
		return false, nil
	}
	if !canConvertToTimeType(srcIndirectValue.Type()) {
		return false, nil
	}

	t, ok := srcIndirectValue.Interface().(time.Time)
	if !ok {
		return false, nil
	}

	// 东八时区
	ts := t.In(time.FixedZone("CST", 8*3600)).Format(formatString)
	rts := reflect.ValueOf(ts)

	if rts.Type().AssignableTo(dstFieldValue.Type()) {
		dstFieldValue.Set(rts)
		return true, nil
	}

	return false, nil
}

func (dc *DeepCopier) processSqlField(tags Tags, srcFieldValue, dstFieldValue reflect.Value) (bool, error) {
	if canConvertToValuerType(srcFieldValue.Type()) {
		// Valuer -> ptr
		if dstFieldValue.Kind() == reflect.Ptr {
			if srcFieldValue.Type().AssignableTo(dstFieldValue.Type()) {
				dstFieldValue.Set(srcFieldValue)
				return true, nil
			}

			valuerValue, _ := srcFieldValue.Interface().(driver.Valuer).Value()
			if valuerValue == nil {
				return false, nil
			}
			valuerType := reflect.TypeOf(valuerValue)

			// new一个valuer类型指针
			ptr := reflect.New(valuerType)
			ptr.Elem().Set(reflect.ValueOf(valuerValue))

			if valuerType.AssignableTo(dstFieldValue.Type().Elem()) {
				dstFieldValue.Set(ptr)
				return true, nil
			}
		} else { // Valuer -> value
			if srcFieldValue.Type().AssignableTo(dstFieldValue.Type()) {
				dstFieldValue.Set(srcFieldValue)
				return true, nil
			}

			valuerValue, _ := srcFieldValue.Interface().(driver.Valuer).Value()
			if valuerValue == nil {
				return false, nil
			}

			rv := reflect.ValueOf(valuerValue)

			// 自定义null.Int判断
			if dstFieldValue.Type().Kind() == reflect.Int && rv.Type().Kind() == reflect.Int64 {
				dstFieldValue.Set(reflect.ValueOf(int(valuerValue.(int64))))
				return true, nil
			}

			if rv.Type().AssignableTo(dstFieldValue.Type()) {
				dstFieldValue.Set(rv)
				return true, nil
			}
		}

		return false, nil
	} else if canConvertToValuerType(dstFieldValue.Type()) { // gopkg.in/guregu/null.v3 转换
		if srcFieldValue.Type().AssignableTo(dstFieldValue.Type()) {
			dstFieldValue.Set(srcFieldValue)
			return true, nil
		}

		srcInterface := srcFieldValue.Interface()
		if srcInterface == nil {
			return false, nil
		}

		dstFieldType := dstFieldValue.Type()
		ptr := reflect.New(dstFieldType)

		initMethod := ptr.MethodByName("SetValid")
		if !initMethod.IsValid() {
			return false, errors.New(fmt.Sprintf("method %s is invalid", initMethod))
		}

		args := []reflect.Value{srcFieldValue}
		initMethod.Call(args)

		if dstFieldValue.Kind() == reflect.Ptr {
			dstFieldValue.Set(ptr)
		} else {
			dstFieldValue.Set(ptr.Elem())
		}

		return true, nil
	} else {
		return false, nil
	}
}

func (dc *DeepCopier) processOthers(tags Tags, srcFieldValue, dstFieldValue reflect.Value) (bool, error) {
	// Ptr -> Value
	if srcFieldValue.Type().Kind() == reflect.Ptr && !srcFieldValue.IsNil() && dstFieldValue.Type().Kind() != reflect.Ptr {
		indirect := reflect.Indirect(srcFieldValue)

		if indirect.Type().AssignableTo(dstFieldValue.Type()) {
			dstFieldValue.Set(indirect)
			return true, nil
		}
	}

	// Value -> Ptr
	if srcFieldValue.Type().Kind() != reflect.Ptr && dstFieldValue.Type().Kind() == reflect.Ptr {
		// new一个指针
		ptr := reflect.New(srcFieldValue.Type())
		ptr.Elem().Set(srcFieldValue)

		if srcFieldValue.Type().AssignableTo(dstFieldValue.Type().Elem()) {
			dstFieldValue.Set(ptr)
			return true, nil
		}
	}

	// Other types
	if srcFieldValue.Type().AssignableTo(dstFieldValue.Type()) {
		dstFieldValue.Set(srcFieldValue)
		return true, nil
	}

	return false, nil
}

// field name对象池，优化内存分配
var fieldNamePool = sync.Pool{
	New: func() interface{} {
		return make([]string, 0)
	},
}

func cleanFieldNameSlice(fieldNames []string) {
	fieldNamePool.Put(fieldNames[0:0])
}

// 反射map对象池，优化内存分配
var reflectValueMapPool = sync.Pool{
	New: func() interface{} {
		return map[string][]reflect.Value{}
	},
}

func cleanReflectValueMap(kv map[string][]reflect.Value) {
	for k, v := range kv {
		v = v[0:0]
		delete(kv, k)
	}
	reflectValueMapPool.Put(kv)
}
