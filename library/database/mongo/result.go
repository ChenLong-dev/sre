package mongo

import (
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

// 单个查询结果
type SingleResult struct {
	*mongo.SingleResult
	err error
}

// 将单个查询结果解码至value中
func (sr *SingleResult) Decode(value interface{}) error {
	if sr.err != nil {
		return sr.err
	}
	return sr.SingleResult.Decode(value)
}

// 查询结果
type FindResult struct {
	raws []bson.Raw
	err  error
}

// 通过bson raw新建结构体
func (fr *FindResult) newDstStructByBsonRaw(raw bson.Raw, dstType reflect.Type) (interface{}, error) {
	vr := bsonrw.NewBSONDocumentReader(raw)
	dec, err := bson.NewDecoder(vr)
	if err != nil {
		return nil, err
	}

	var item interface{}
	switch dstType.Kind() {
	case reflect.Slice:
		item = reflect.New(dstType.Elem()).Interface()
	case reflect.Struct:
		item = reflect.New(dstType).Interface()
	default:
		return nil, errors.New("value is not support type")
	}

	err = dec.Decode(item)
	if err != nil {
		return nil, err
	}

	return item, nil
}

// 解码切片
func (fr *FindResult) decodeSlice(dst interface{}) (err error) {
	dstType := reflect.TypeOf(dst).Elem()
	dstValue := reflect.ValueOf(dst).Elem()
	if dstType.Kind() != reflect.Slice {
		return errors.New("value is not slice")
	}

	for _, raw := range fr.raws {
		item, err := fr.newDstStructByBsonRaw(raw, dstType)
		if err != nil {
			return err
		}

		dstValue = reflect.Append(dstValue, reflect.ValueOf(item).Elem())
	}
	reflect.ValueOf(dst).Elem().Set(dstValue)

	return
}

// 解码结构体
func (fr *FindResult) decodeStruct(dst interface{}) (err error) {
	dstType := reflect.TypeOf(dst).Elem()
	if dstType.Kind() != reflect.Struct {
		return errors.New("value is not struct")
	}

	if len(fr.raws) < 1 {
		return nil
	}

	item, err := fr.newDstStructByBsonRaw(fr.raws[0], dstType)
	if err != nil {
		return err
	}

	reflect.ValueOf(dst).Elem().Set(reflect.ValueOf(item).Elem())

	return
}

// 将查询结果解码至value中
func (fr *FindResult) Decode(value interface{}) (err error) {
	if fr.err != nil {
		return fr.err
	}
	if value == nil {
		return errors.New("value is nil")
	}
	if reflect.TypeOf(value).Kind() != reflect.Ptr {
		return errors.New("value is not ptr")
	}

	switch reflect.TypeOf(value).Elem().Kind() {
	case reflect.Slice:
		return fr.decodeSlice(value)
	case reflect.Struct:
		return fr.decodeStruct(value)
	default:
		return errors.New("value is not support type")
	}
}
