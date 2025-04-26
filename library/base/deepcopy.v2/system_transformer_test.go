package deepcopy

import (
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.shanhai.int/sre/library/base/null"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// sql类型转换
type sqlSrc struct {
	IntToSql    int
	BoolToSql   bool
	FloatToSql  float64
	StringToSql string
	Int64ToSql  int64
	TimeToSql   time.Time

	SqlToInt       null.Int
	SqlToBool      null.Bool
	SqlToFloat     null.Float
	SqlToString    null.String
	SqlToInt64     null.Int64
	SqlToTime      null.Time
	SqlToInterface null.String

	MapToStruct map[string]interface{}
	MapToPtr    map[string]string

	StructToMap sqlChildSrc
	PtrToMap    *sqlChildSrc

	SliceStringToSql []string
	SliceSqlToString []null.String

	MapStringToSql map[string]null.String
	MapSqlToString map[string]string
}
type sqlChildSrc struct {
	String null.String
	Int    int
}
type sqlDst struct {
	IntToSql    null.Int
	BoolToSql   null.Bool
	FloatToSql  null.Float
	StringToSql null.String
	Int64ToSql  null.Int64
	TimeToSql   null.Time

	SqlToInt       int
	SqlToBool      bool
	SqlToFloat     float64
	SqlToString    string
	SqlToInt64     int64
	SqlToTime      time.Time
	SqlToInterface interface{}

	MapToStruct sqlChildSrc
	MapToPtr    *sqlChildSrc

	StructToMap map[string]interface{}
	PtrToMap    map[string]string

	SliceStringToSql []null.String
	SliceSqlToString []string

	MapStringToSql map[string]string
	MapSqlToString map[string]null.String
}

var sqlTest = []TestCase{
	{
		Name: "sql-dst",
		Src: &sqlSrc{
			IntToSql:    1,
			BoolToSql:   true,
			FloatToSql:  2.3,
			StringToSql: "abc",
			Int64ToSql:  2,
			TimeToSql:   time.Date(2015, 4, 25, 22, 30, 15, 0, time.Local),
		},
		Dst: new(sqlDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*sqlSrc)
			dst := dc.dst.(*sqlDst)
			assert.Nil(t, err)
			assert.Equal(t, src.IntToSql, dst.IntToSql.ValueOrZero())
			assert.Equal(t, src.BoolToSql, dst.BoolToSql.ValueOrZero())
			assert.Equal(t, src.FloatToSql, dst.FloatToSql.ValueOrZero())
			assert.Equal(t, src.StringToSql, dst.StringToSql.ValueOrZero())
			assert.Equal(t, src.Int64ToSql, dst.Int64ToSql.ValueOrZero())
			assert.Equal(t, src.TimeToSql, dst.TimeToSql.ValueOrZero())
		},
	},
	{
		Name: "sql-to",
		Src: &sqlSrc{
			SqlToInt:       null.IntFrom(1),
			SqlToBool:      null.BoolFrom(true),
			SqlToFloat:     null.FloatFrom(2.3),
			SqlToString:    null.StringFrom("abc"),
			SqlToInterface: null.StringFrom("123"),
			SqlToInt64:     null.Int64From(2),
			SqlToTime:      null.TimeFrom(time.Date(2015, 4, 25, 22, 30, 15, 0, time.Local)),
		},
		Dst: new(sqlDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*sqlSrc)
			dst := dc.dst.(*sqlDst)
			assert.Nil(t, err)
			assert.Equal(t, src.SqlToInt.ValueOrZero(), dst.SqlToInt)
			assert.Equal(t, src.SqlToBool.ValueOrZero(), dst.SqlToBool)
			assert.Equal(t, src.SqlToFloat.ValueOrZero(), dst.SqlToFloat)
			assert.Equal(t, src.SqlToString.ValueOrZero(), dst.SqlToString)
			assert.Equal(t, src.SqlToInterface.ValueOrZero(), dst.SqlToInterface.(string))
			assert.Equal(t, src.SqlToInt64.ValueOrZero(), dst.SqlToInt64)
			assert.Equal(t, src.SqlToTime.ValueOrZero(), dst.SqlToTime)
		},
	},
	{
		Name: "sql-map_to_struct",
		Src: &sqlSrc{
			MapToStruct: map[string]interface{}{
				"Int":    null.IntFrom(1),
				"String": "abc",
			},
		},
		Dst: new(sqlDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*sqlSrc)
			dst := dc.dst.(*sqlDst)
			assert.Nil(t, err)
			assert.Equal(t, src.MapToStruct["Int"].(null.Int).ValueOrZero(), dst.MapToStruct.Int)
			assert.Equal(t, src.MapToStruct["String"].(string), dst.MapToStruct.String.ValueOrZero())
		},
	},
	{
		Name: "sql-map_to_ptr",
		Src: &sqlSrc{
			MapToPtr: map[string]string{
				"String": "abc",
			},
		},
		Dst: new(sqlDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*sqlSrc)
			dst := dc.dst.(*sqlDst)
			assert.Nil(t, err)
			assert.Equal(t, src.MapToPtr["String"], dst.MapToPtr.String.ValueOrZero())
		},
	},
	{
		Name: "sql-struct_to_map",
		Src: &sqlSrc{
			StructToMap: sqlChildSrc{
				Int:    123,
				String: null.StringFrom("abc"),
			},
		},
		Dst: new(sqlDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*sqlSrc)
			dst := dc.dst.(*sqlDst)
			assert.Nil(t, err)
			assert.Equal(t, src.StructToMap.Int, dst.StructToMap["Int"].(int))
			assert.Equal(t, src.StructToMap.String.ValueOrZero(), dst.StructToMap["String"].(null.String).ValueOrZero())
		},
	},
	{
		Name: "sql-ptr_to_map",
		Src: &sqlSrc{
			PtrToMap: &sqlChildSrc{
				Int:    123,
				String: null.StringFrom("abc"),
			},
		},
		Dst: new(sqlDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*sqlSrc)
			dst := dc.dst.(*sqlDst)
			assert.Nil(t, err)
			assert.Equal(t, src.PtrToMap.String.ValueOrZero(), dst.PtrToMap["String"])
		},
	},
	{
		Name: "sql-slice_dst",
		Src: &sqlSrc{
			SliceStringToSql: []string{"a", "123"},
		},
		Dst: new(sqlDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*sqlSrc)
			dst := dc.dst.(*sqlDst)
			assert.Nil(t, err)
			assert.Equal(t, len(src.SliceStringToSql), len(dst.SliceStringToSql))
			assert.Equal(t, src.SliceStringToSql[0], dst.SliceStringToSql[0].ValueOrZero())
			assert.Equal(t, src.SliceStringToSql[1], dst.SliceStringToSql[1].ValueOrZero())
		},
	},
	{
		Name: "sql-slice_to",
		Src: &sqlSrc{
			SliceSqlToString: []null.String{
				null.StringFrom("a"),
				null.StringFrom("123"),
			},
		},
		Dst: new(sqlDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*sqlSrc)
			dst := dc.dst.(*sqlDst)
			assert.Nil(t, err)
			assert.Equal(t, len(src.SliceSqlToString), len(dst.SliceSqlToString))
			assert.Equal(t, src.SliceSqlToString[0].ValueOrZero(), dst.SliceSqlToString[0])
			assert.Equal(t, src.SliceSqlToString[1].ValueOrZero(), dst.SliceSqlToString[1])
		},
	},
	{
		Name: "sql-map_dst",
		Src: &sqlSrc{
			MapSqlToString: map[string]string{
				"0": "a",
				"1": "123",
			},
		},
		Dst: new(sqlDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*sqlSrc)
			dst := dc.dst.(*sqlDst)
			assert.Nil(t, err)
			assert.Equal(t, len(src.MapSqlToString), len(dst.MapSqlToString))
			assert.Equal(t, src.MapSqlToString["0"], dst.MapSqlToString["0"].ValueOrZero())
			assert.Equal(t, src.MapSqlToString["1"], dst.MapSqlToString["1"].ValueOrZero())
		},
	},
	{
		Name: "sql-map_to",
		Src: &sqlSrc{
			MapStringToSql: map[string]null.String{
				"0": null.StringFrom("a"),
				"1": null.StringFrom("123"),
			},
		},
		Dst: new(sqlDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*sqlSrc)
			dst := dc.dst.(*sqlDst)
			assert.Nil(t, err)
			assert.Equal(t, len(src.MapStringToSql), len(dst.MapStringToSql))
			assert.Equal(t, src.MapStringToSql["0"].ValueOrZero(), dst.MapStringToSql["0"])
			assert.Equal(t, src.MapStringToSql["1"].ValueOrZero(), dst.MapStringToSql["1"])
		},
	},
}

// timeformat tag
type timeFormatTagSrc struct {
	Time          time.Time
	TimePtr       *time.Time
	TimeInterface interface{}
	Map           map[string]interface{}
}
type timeFormatConvertErrorSrc struct {
	Time int
}
type timeFormatTagDst struct {
	Time          string `deepcopy:"timeformat:2006-01-02 15:04:05"`
	TimePtr       string `deepcopy:"timeformat:2006-01-02 15:04:05"`
	TimeInterface string `deepcopy:"timeformat:2006-01-02 15:04:05"`
	Map           timeFormatTagChildDst
}
type timeFormatTagChildDst struct {
	ChildTime    string `deepcopy:"timeformat:2006-01-02 15:04:05"`
	ChildTimePtr string `deepcopy:"timeformat:2006-01-02 15:04:05"`
}

var timeFormatTagTime = time.Date(2015, 4, 25, 22, 30, 15, 0, time.Local)
var timeFormatTagTest = []TestCase{
	{
		Name: "timeformat-struct",
		Src: &timeFormatTagSrc{
			Time: timeFormatTagTime,
		},
		Dst: new(timeFormatTagDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*timeFormatTagSrc)
			dst := dc.dst.(*timeFormatTagDst)
			assert.Nil(t, err)
			assert.Equal(t, src.Time.Format("2006-01-02 15:04:05"), dst.Time)
		},
	},
	{
		Name: "timeformat-ptr",
		Src: &timeFormatTagSrc{
			TimePtr: &timeFormatTagTime,
		},
		Dst: new(timeFormatTagDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*timeFormatTagSrc)
			dst := dc.dst.(*timeFormatTagDst)
			assert.Nil(t, err)
			assert.Equal(t, src.TimePtr.Format("2006-01-02 15:04:05"), dst.TimePtr)
		},
	},
	{
		Name: "timeformat-ptr_interface",
		Src: &timeFormatTagSrc{
			TimeInterface: &timeFormatTagTime,
		},
		Dst: new(timeFormatTagDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*timeFormatTagSrc)
			dst := dc.dst.(*timeFormatTagDst)
			assert.Nil(t, err)
			assert.Equal(t, src.TimeInterface.(*time.Time).Format("2006-01-02 15:04:05"), dst.TimeInterface)
		},
	},
	{
		Name: "timeformat-map_to_struct",
		Src: &timeFormatTagSrc{
			Map: map[string]interface{}{
				"ChildTime": timeFormatTagTime,
			},
		},
		Dst: new(timeFormatTagDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*timeFormatTagSrc)
			dst := dc.dst.(*timeFormatTagDst)
			assert.Nil(t, err)
			assert.Equal(t, src.Map["ChildTime"].(time.Time).Format("2006-01-02 15:04:05"), dst.Map.ChildTime)
		},
	},
	{
		Name: "timeformat-map_to_ptr",
		Src: &timeFormatTagSrc{
			Map: map[string]interface{}{
				"ChildTimePtr": &timeFormatTagTime,
			},
		},
		Dst: new(timeFormatTagDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*timeFormatTagSrc)
			dst := dc.dst.(*timeFormatTagDst)
			assert.Nil(t, err)
			assert.Equal(t, src.Map["ChildTimePtr"].(*time.Time).Format("2006-01-02 15:04:05"), dst.Map.ChildTimePtr)
		},
	},
	{
		Name: "timeformat-convert_error_normal",
		Src:  &timeFormatConvertErrorSrc{},
		Dst:  new(timeFormatTagDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			assert.Nil(t, err)
		},
	},
	{
		Name: "timeformat-convert_error_strict",
		Src:  &timeFormatConvertErrorSrc{},
		Dst:  new(timeFormatTagDst),
		Config: &Config{
			StrictMode: true,
		},
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			assert.NotNil(t, err)
		},
	},
}

// string tag
type stringTagSrc struct {
	Int    int
	Bool   bool
	Float  float64
	String string

	Time    time.Time
	TimePtr *time.Time

	stringTagChildSrc
	Struct stringTagChildSrc
	Ptr    *stringTagChildSrc

	NullInt    null.Int
	NullBool   null.Bool
	NullFloat  null.Float
	NullString null.String
	NullInt64  null.Int64
	NullTime   null.Time
}
type stringTagChildSrc struct {
	ChildString string
}
type stringTagDst struct {
	Int    string `deepcopy:"string"`
	Bool   string `deepcopy:"string"`
	Float  string `deepcopy:"string"`
	String string `deepcopy:"string"`
}
type stringTagTimeDst struct {
	Time    string `deepcopy:"string"`
	TimePtr string `deepcopy:"string"`
}
type stringTagStructDst struct {
	ChildString string `deepcopy:"string"`
	Struct      string `deepcopy:"string"`
	Ptr         string `deepcopy:"string"`
}
type stringTagNullDst struct {
	NullInt    string `deepcopy:"string"`
	NullBool   string `deepcopy:"string"`
	NullFloat  string `deepcopy:"string"`
	NullString string `deepcopy:"string"`
	NullInt64  string `deepcopy:"string"`
	NullTime   string `deepcopy:"string"`
}
type stringTagErrorDst struct {
	Int int `deepcopy:"string"`
}

var stringTagTime = time.Date(2015, 4, 25, 22, 30, 15, 0, time.Local)
var stringTagTest = []TestCase{
	{
		Name: "string-base",
		Src: &stringTagSrc{
			Int:    1,
			Bool:   true,
			Float:  2.3,
			String: "abc",
		},
		Dst: new(stringTagDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			dst := dc.dst.(*stringTagDst)
			assert.Nil(t, err)
			assert.Equal(t, "1", dst.Int)
			assert.Equal(t, "true", dst.Bool)
			assert.Equal(t, "2.3", dst.Float)
			assert.Equal(t, "abc", dst.String)
		},
	},
	{
		Name: "string-time",
		Src: &stringTagSrc{
			Time:    stringTagTime,
			TimePtr: &stringTagTime,
		},
		Dst: new(stringTagTimeDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			dst := dc.dst.(*stringTagTimeDst)
			assert.Nil(t, err)
			assert.Equal(t, "2015-04-25 22:30:15 +0800 CST", dst.Time)
			assert.Equal(t, "2015-04-25 22:30:15 +0800 CST", dst.TimePtr)
		},
	},
	{
		Name: "string-struct",
		Src: &stringTagSrc{
			stringTagChildSrc: stringTagChildSrc{
				ChildString: "anonymous",
			},
			Struct: stringTagChildSrc{
				ChildString: "struct",
			},
			Ptr: &stringTagChildSrc{
				ChildString: "ptr",
			},
		},
		Dst: new(stringTagStructDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			dst := dc.dst.(*stringTagStructDst)
			assert.Nil(t, err)
			assert.Equal(t, "anonymous", dst.ChildString)
			assert.Equal(t, "{struct}", dst.Struct)
			assert.Equal(t, "&{ptr}", dst.Ptr)
		},
	},
	{
		Name: "string-null",
		Src: &stringTagSrc{
			NullInt:    null.IntFrom(1),
			NullBool:   null.BoolFrom(true),
			NullFloat:  null.FloatFrom(1.23),
			NullString: null.StringFrom("abc"),
			NullInt64:  null.Int64From(2),
			NullTime:   null.TimeFrom(stringTagTime),
		},
		Dst: new(stringTagNullDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			dst := dc.dst.(*stringTagNullDst)
			assert.Nil(t, err)
			assert.Equal(t, "1", dst.NullInt)
			assert.Equal(t, "true", dst.NullBool)
			assert.Equal(t, "1.23", dst.NullFloat)
			assert.Equal(t, "abc", dst.NullString)
			assert.Equal(t, "2", dst.NullInt64)
			assert.Equal(t, "2015-04-25 22:30:15 +0800 CST", dst.NullTime)
		},
	},
	{
		Name: "string-error_normal",
		Src: &stringTagSrc{
			Int: 1,
		},
		Dst: new(stringTagErrorDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			assert.Nil(t, err)
		},
	},
	{
		Name: "string-error_strict",
		Src: &stringTagSrc{
			Int: 1,
		},
		Dst: new(stringTagErrorDst),
		Config: &Config{
			StrictMode: true,
		},
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			assert.NotNil(t, err)
		},
	},
}

// bool tag
type boolTagSrc struct {
	Int       int
	IntFalse  int
	Int64     int64
	UInt      uint
	UInt64    uint64
	Float     float64
	NullInt   null.Int
	NullInt64 null.Int64
	NullFloat null.Float
}
type boolTagDst struct {
	Int      bool `deepcopy:"bool"`
	IntFalse bool `deepcopy:"bool"`
	Int64    bool `deepcopy:"bool"`
	UInt     bool `deepcopy:"bool"`
	UInt64   bool `deepcopy:"bool"`
	Float    bool `deepcopy:"bool"`
}
type boolTagNullDst struct {
	NullInt   bool `deepcopy:"bool"`
	NullInt64 bool `deepcopy:"bool"`
	NullFloat bool `deepcopy:"bool"`
}
type boolTagErrorDst struct {
	Int int `deepcopy:"bool"`
}

var boolTagTest = []TestCase{
	{
		Name: "bool-base",
		Src: &boolTagSrc{
			Int:      1,
			IntFalse: 0,
			Int64:    2,
			UInt:     3,
			UInt64:   4,
			Float:    1.23,
		},
		Dst: new(boolTagDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			dst := dc.dst.(*boolTagDst)
			assert.Nil(t, err)
			assert.Equal(t, true, dst.Int)
			assert.Equal(t, false, dst.IntFalse)
			assert.Equal(t, true, dst.Int64)
			assert.Equal(t, true, dst.UInt)
			assert.Equal(t, true, dst.UInt64)
			assert.Equal(t, true, dst.Float)
		},
	},
	{
		Name: "bool-null",
		Src: &boolTagSrc{
			NullInt:   null.IntFrom(1),
			NullInt64: null.Int64From(2),
			NullFloat: null.FloatFrom(1.23),
		},
		Dst: new(boolTagNullDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			dst := dc.dst.(*boolTagNullDst)
			assert.Nil(t, err)
			assert.Equal(t, true, dst.NullInt)
			assert.Equal(t, true, dst.NullInt64)
			assert.Equal(t, true, dst.NullFloat)
		},
	},
	{
		Name: "bool-error_normal",
		Src: &boolTagSrc{
			Int: 1,
		},
		Dst: new(boolTagErrorDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			assert.Nil(t, err)
		},
	},
	{
		Name: "bool-error_strict",
		Src: &boolTagSrc{
			Int: 1,
		},
		Dst: new(boolTagErrorDst),
		Config: &Config{
			StrictMode: true,
		},
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			assert.NotNil(t, err)
		},
	},
}

// objectId tag
type objectIDTagSrc struct {
	ObjectID          primitive.ObjectID
	ObjectIDPtr       *primitive.ObjectID
	ObjectIDInterface interface{}

	ObjectString    string
	ObjectStringPtr string
}

type objectIDTagNullSrc struct {
	ObjectString string
}

type objectIDTagConvertErrorSrc struct {
	ObjectID int
}
type objectIDTagDst struct {
	ObjectID          string `deepcopy:"objectid"`
	ObjectIDPtr       string `deepcopy:"objectid"`
	ObjectIDInterface string `deepcopy:"objectid"`

	ObjectString    primitive.ObjectID  `deepcopy:"objectid"`
	ObjectStringPtr *primitive.ObjectID `deepcopy:"objectid"`
}

var objectIDTagObjectID = primitive.NewObjectID()
var objectIDTagString = primitive.NewObjectID().Hex()

var objectIDTagTest = []TestCase{
	{
		Name: "objectid-struct_to_string",
		Src: &objectIDTagSrc{
			ObjectID: objectIDTagObjectID,
		},
		Dst: new(objectIDTagDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			dst := dc.dst.(*objectIDTagDst)
			src := dc.src.(*objectIDTagSrc)
			assert.Nil(t, err)
			assert.Equal(t, src.ObjectID.Hex(), dst.ObjectID)
		},
	},
	{
		Name: "objectid-ptr_to_string",
		Src: &objectIDTagSrc{
			ObjectIDPtr: &objectIDTagObjectID,
		},
		Dst: new(objectIDTagDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			dst := dc.dst.(*objectIDTagDst)
			src := dc.src.(*objectIDTagSrc)
			assert.Nil(t, err)
			assert.Equal(t, src.ObjectIDPtr.Hex(), dst.ObjectIDPtr)
		},
	},
	{
		Name: "objectid-ptr_interface_to_string",
		Src: &objectIDTagSrc{
			ObjectIDInterface: &objectIDTagObjectID,
		},
		Dst: new(objectIDTagDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			dst := dc.dst.(*objectIDTagDst)
			src := dc.src.(*objectIDTagSrc)
			assert.Nil(t, err)
			assert.Equal(t, src.ObjectIDInterface.(*primitive.ObjectID).Hex(), dst.ObjectIDInterface)
		},
	},
	{
		Name: "objectid-string_to_objectid",
		Src: &objectIDTagSrc{
			ObjectString: objectIDTagString,
		},
		Dst: new(objectIDTagDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			dst := dc.dst.(*objectIDTagDst)
			src := dc.src.(*objectIDTagSrc)
			assert.Nil(t, err)
			objectID, err := primitive.ObjectIDFromHex(src.ObjectString)
			assert.Nil(t, err)
			assert.Equal(t, objectID, dst.ObjectString)
		},
	},
	{
		Name: "objectid-string_to_ptr",
		Src: &objectIDTagSrc{
			ObjectStringPtr: objectIDTagString,
		},
		Dst: new(objectIDTagDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			dst := dc.dst.(*objectIDTagDst)
			src := dc.src.(*objectIDTagSrc)
			assert.Nil(t, err)
			objectID, err := primitive.ObjectIDFromHex(src.ObjectStringPtr)
			assert.Nil(t, err)
			assert.Equal(t, objectID, *dst.ObjectStringPtr)
		},
	},
	{
		Name: "objectid-nullStr_to_objectid_strict",
		Src: &objectIDTagNullSrc{
			ObjectString: "",
		},
		Dst: new(objectIDTagDst),
		Config: &Config{
			StrictMode: true,
		},
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			assert.NotNil(t, err)
		},
	},
	{
		Name: "objectid-nullStr_to_objectid",
		Src: &objectIDTagNullSrc{
			ObjectString: "",
		},
		Dst: new(objectIDTagDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			dst := dc.dst.(*objectIDTagDst)
			assert.Nil(t, err)

			assert.Equal(t, primitive.NilObjectID, dst.ObjectString)
		},
	},

	{
		Name: "objectid-convert_error_normal",
		Src: &objectIDTagConvertErrorSrc{
			ObjectID: 1,
		},
		Dst: new(objectIDTagDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			assert.Nil(t, err)
		},
	},
	{
		Name: "objectid-convert_error_strict",
		Src: &objectIDTagConvertErrorSrc{
			ObjectID: 1,
		},
		Dst: new(objectIDTagDst),
		Config: &Config{
			StrictMode: true,
		},
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			assert.NotNil(t, err)
		},
	},
}
