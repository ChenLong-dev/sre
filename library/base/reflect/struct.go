package reflect

import (
	"bytes"
	"github.com/BurntSushi/toml"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"reflect"
)

// 结构体转字典
func StructToMap(obj interface{}) (map[string]interface{}, error) {
	return StructToMapByJson(obj)
}

// 结构体转字典，json实现
func StructToMapByJson(obj interface{}) (map[string]interface{}, error) {
	json := jsoniter.ConfigCompatibleWithStandardLibrary

	m := make(map[string]interface{})
	jsonString, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonString, &m)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// 字典转结构体，toml实现
func MapToStructByToml(obj map[string]interface{}, dest interface{}) error {
	buf := new(bytes.Buffer)

	err := toml.NewEncoder(buf).Encode(obj)
	if err != nil {
		return err
	}

	err = toml.Unmarshal(buf.Bytes(), dest)
	if err != nil {
		return err
	}

	return nil
}

// 结构体转字典，反射实现
func StructToMapByReflect(obj interface{}) map[string]interface{} {
	v := reflect.Indirect(reflect.ValueOf(obj))
	t := v.Type()

	var data = make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		var (
			vField = v.Field(i)
			tField = v.Type().Field(i)
		)

		if tField.Anonymous {
			childMap := StructToMapByReflect(vField.Addr().Interface())
			for childKey, childValue := range childMap {
				data[childKey] = childValue
			}
		} else {
			data[tField.Name] = vField.Interface()
		}
	}
	return data
}

// interface转interface数组，反射实现
func InterfaceToSlice(obj interface{}) ([]interface{}, error) {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Slice {
		return nil, errors.New("obj type is not slice")
	}

	l := v.Len()
	res := make([]interface{}, l)
	for i := 0; i < l; i++ {
		res[i] = v.Index(i).Interface()
	}

	return res, nil
}
