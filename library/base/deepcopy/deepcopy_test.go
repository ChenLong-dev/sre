package deepcopy

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.shanhai.int/sre/library/base/null"
)

type TestBeanSrc struct {
	TestInt       int
	TestBool      bool
	TestString    string
	TestInterface interface{}

	TestSrcMethod     int
	TestSrcFrom       string
	TestSrcTo         float64 `deepcopy:"to:TestDstTo"`
	TestSkip          string
	TestForce         *time.Time
	TestSql           sql.NullInt64
	TestTimeFormatPtr *time.Time
	TestTimeFormat    time.Time

	TestAnonymousValueInt int
	TestAnonymousValue
	*TestAnonymousPtr

	TestStruct           TestStruct
	TestStructPtr        *TestStruct
	TestStructPtrToValue *TestStruct
	TestStructValueToPtr TestStruct
}

func (src *TestBeanSrc) GenerateMethodResult(m map[string]interface{}) string {
	return m["time"].(time.Time).Format("2006/01/02")
}

func (src *TestBeanSrc) GenerateMethodPtrResult(m map[string]interface{}) *TestStruct {
	return &TestStruct{
		TestStructInt: m["time"].(time.Time).Nanosecond(),
	}
}

func (src *TestBeanSrc) GenerateMethodValueResult(m map[string]interface{}) TestStruct {
	return TestStruct{
		TestStructInt: m["time"].(time.Time).Nanosecond(),
	}
}

type TestBeanDst struct {
	TestInt       int
	TestBool      bool
	TestString    string
	TestInterface interface{}

	TestDstMethod      string      `deepcopy:"method:GenerateMethodResult"`
	TestDstPtrMethod   TestStruct  `deepcopy:"method:GenerateMethodPtrResult"`
	TestDstValueMethod *TestStruct `deepcopy:"method:GenerateMethodValueResult"`
	TestDstFrom        string      `deepcopy:"from:TestSrcFrom"`
	TestDstTo          float64
	TestSkip           string      `deepcopy:"skip"`
	TestForce          interface{} `deepcopy:"force"`
	TestSql            int64       `deepcopy:"sql"`
	TestTimeFormatPtr  string      `deepcopy:"timeformat:2006/01/02 15:04:05"`
	TestTimeFormat     string      `deepcopy:"timeformat:2006/01/02 15:04:05"`

	TestAnonymousValueInt int
	TestAnonymousValue
	*TestAnonymousPtr

	TestStruct           TestStruct
	TestStructPtr        *TestStruct
	TestStructPtrToValue TestStruct
	TestStructValueToPtr *TestStruct
}

type TestAnonymousPtr struct {
	TestAnonymousPtrInt int
}

type TestAnonymousValue struct {
	TestAnonymousValueInt int
}

type TestStruct struct {
	TestStructInt int
}

type TestBeanAnonymousSrcA struct {
	TestAnonymousValueInt int
	TestAnonymousValue
}

type TestBeanAnonymousDstA struct {
	TestAnonymousValueInt int
	TestAnonymousValue
}

type TestBeanAnonymousSrcB struct {
	TestAnonymousValue
}

type TestBeanAnonymousDstB struct {
	TestAnonymousValueInt int
	TestAnonymousValue
}

type TestBeanAnonymousSrcC struct {
	TestAnonymousValueInt int
	TestAnonymousValue
}

type TestBeanAnonymousDstC struct {
	TestAnonymousValue
}

type TestBeanAnonymousSrcD struct {
	TestAnonymousValue
}

type TestBeanAnonymousDstD struct {
	TestAnonymousValue
}

func TestDeepCopyStruct(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		nowTime := time.Now()

		src := &TestBeanSrc{
			TestInt:       101,
			TestBool:      true,
			TestString:    "this is a test",
			TestInterface: "interface",

			TestSrcMethod: 1,
			TestSrcFrom:   "11",
			TestSrcTo:     1.1,
			TestSkip:      "22",
			TestForce:     &nowTime,
			TestSql: sql.NullInt64{
				Int64: 33,
				Valid: true,
			},
			TestTimeFormatPtr: &nowTime,
			TestTimeFormat:    nowTime,

			TestStruct: TestStruct{
				TestStructInt: 2,
			},
			TestStructPtr: &TestStruct{
				TestStructInt: 4,
			},
			TestStructPtrToValue: &TestStruct{
				TestStructInt: 7,
			},
			TestStructValueToPtr: TestStruct{
				TestStructInt: 8,
			},

			TestAnonymousValue: TestAnonymousValue{
				TestAnonymousValueInt: 5,
			},
			TestAnonymousValueInt: 99,
			TestAnonymousPtr: &TestAnonymousPtr{
				TestAnonymousPtrInt: 6,
			},
		}
		dst := &TestBeanDst{}

		args := make(map[string]interface{})
		args["time"] = nowTime

		if err := Copy(src).WithArgs(args).To(dst); err != nil {
			t.Fatalf("%#v\n", err)
		}

		assert.Equal(t, src.TestInt, dst.TestInt)
		assert.Equal(t, src.TestBool, dst.TestBool)
		assert.Equal(t, src.TestString, dst.TestString)
		assert.Equal(t, src.TestInterface, dst.TestInterface)
		assert.Equal(t, src.GenerateMethodResult(args), dst.TestDstMethod)
		assert.Equal(t, src.GenerateMethodPtrResult(args).TestStructInt, dst.TestDstPtrMethod.TestStructInt)
		assert.Equal(t, src.GenerateMethodValueResult(args).TestStructInt, dst.TestDstValueMethod.TestStructInt)
		assert.Equal(t, src.TestSrcFrom, dst.TestDstFrom)
		assert.Equal(t, src.TestSrcTo, dst.TestDstTo)
		assert.Equal(t, "", dst.TestSkip)
		assert.Equal(t, src.TestForce, dst.TestForce)
		assert.Equal(t, src.TestSql.Int64, dst.TestSql)
		assert.Equal(t, src.TestTimeFormatPtr.Format("2006/01/02 15:04:05"), dst.TestTimeFormatPtr)
		assert.Equal(t, src.TestTimeFormat.Format("2006/01/02 15:04:05"), dst.TestTimeFormat)
		assert.Equal(t, src.TestStruct.TestStructInt, dst.TestStruct.TestStructInt)
		assert.Equal(t, src.TestStructPtr.TestStructInt, dst.TestStructPtr.TestStructInt)
		assert.Equal(t, src.TestStructPtrToValue.TestStructInt, dst.TestStructPtrToValue.TestStructInt)
		assert.Equal(t, src.TestStructValueToPtr.TestStructInt, dst.TestStructValueToPtr.TestStructInt)
		assert.Equal(t, src.TestAnonymousValue.TestAnonymousValueInt, dst.TestAnonymousValue.TestAnonymousValueInt)
		assert.Equal(t, src.TestAnonymousValueInt, dst.TestAnonymousValueInt)
		assert.Equal(t, src.TestAnonymousPtr.TestAnonymousPtrInt, dst.TestAnonymousPtr.TestAnonymousPtrInt)

	})
}

func TestDeepCopyAnonymousStruct(t *testing.T) {
	t.Run("A", func(t *testing.T) {
		srcA := &TestBeanAnonymousSrcA{
			TestAnonymousValue: TestAnonymousValue{
				TestAnonymousValueInt: 5,
			},
			TestAnonymousValueInt: 99,
		}

		t.Run("ParseAnonymousStruct: false", func(t *testing.T) {
			dstA := &TestBeanAnonymousDstA{}
			if err := Copy(srcA).SetConfig(
				&Config{
					ParseAnonymousStruct: false,
				},
			).To(dstA); err != nil {
				t.Fatalf("%#v\n", err)
			}
			assert.Equal(t, srcA.TestAnonymousValueInt, dstA.TestAnonymousValueInt)
			assert.Equal(t, srcA.TestAnonymousValue.TestAnonymousValueInt, dstA.TestAnonymousValue.TestAnonymousValueInt)

		})
		t.Run("ParseAnonymousStruct: true", func(t *testing.T) {
			dstA := &TestBeanAnonymousDstA{}
			if err := Copy(srcA).SetConfig(
				&Config{
					ParseAnonymousStruct: true,
				},
			).To(dstA); err != nil {
				t.Fatalf("%#v\n", err)
			}
			assert.Equal(t, srcA.TestAnonymousValueInt, dstA.TestAnonymousValueInt)
			// 由于dst解析出匿名结构体，且外部有同名字段，所以内部没有赋值
			assert.Equal(t, 0, dstA.TestAnonymousValue.TestAnonymousValueInt)
		})
	})

	t.Run("B", func(t *testing.T) {
		srcB := &TestBeanAnonymousSrcB{
			TestAnonymousValue: TestAnonymousValue{
				TestAnonymousValueInt: 5,
			},
		}

		t.Run("ParseAnonymousStruct: false", func(t *testing.T) {
			dstB := &TestBeanAnonymousDstB{}
			if err := Copy(srcB).SetConfig(
				&Config{
					ParseAnonymousStruct: false,
				},
			).To(dstB); err != nil {
				t.Fatalf("%#v\n", err)
			}
			assert.Equal(t, srcB.TestAnonymousValue.TestAnonymousValueInt, dstB.TestAnonymousValue.TestAnonymousValueInt)
			// 内部调用FieldByName方法，因为无法判断匿名结构体，所以都会赋值
			assert.Equal(t, srcB.TestAnonymousValue.TestAnonymousValueInt, dstB.TestAnonymousValueInt)

		})
		t.Run("ParseAnonymousStruct: true", func(t *testing.T) {
			dstB := &TestBeanAnonymousDstB{}
			if err := Copy(srcB).SetConfig(
				&Config{
					ParseAnonymousStruct: true,
				},
			).To(dstB); err != nil {
				t.Fatalf("%#v\n", err)
			}
			// 由于dst解析出匿名结构体，且外部有同名字段，所以内部字段没有赋值
			assert.Equal(t, 0, dstB.TestAnonymousValue.TestAnonymousValueInt)
			// 由于dst解析出匿名结构体，且外部有同名字段，所以被赋值
			assert.Equal(t, srcB.TestAnonymousValue.TestAnonymousValueInt, dstB.TestAnonymousValueInt)

		})
	})

	t.Run("C", func(t *testing.T) {
		srcC := &TestBeanAnonymousSrcC{
			TestAnonymousValue: TestAnonymousValue{
				TestAnonymousValueInt: 5,
			},
			TestAnonymousValueInt: 99,
		}

		t.Run("ParseAnonymousStruct: false", func(t *testing.T) {
			dstC := &TestBeanAnonymousDstC{}
			if err := Copy(srcC).SetConfig(
				&Config{
					ParseAnonymousStruct: false,
				},
			).To(dstC); err != nil {
				t.Fatalf("%#v\n", err)
			}
			assert.Equal(t, srcC.TestAnonymousValue.TestAnonymousValueInt, dstC.TestAnonymousValue.TestAnonymousValueInt)
		})
		t.Run("ParseAnonymousStruct: true", func(t *testing.T) {
			dstC := &TestBeanAnonymousDstC{}
			if err := Copy(srcC).SetConfig(
				&Config{
					ParseAnonymousStruct: true,
				},
			).To(dstC); err != nil {
				t.Fatalf("%#v\n", err)
			}
			// 由于dst解析出匿名结构体，但外部无同名字段，所以内部字段被赋值为外部字段
			assert.Equal(t, srcC.TestAnonymousValueInt, dstC.TestAnonymousValue.TestAnonymousValueInt)
			assert.NotEqual(t, srcC.TestAnonymousValue.TestAnonymousValueInt, dstC.TestAnonymousValue.TestAnonymousValueInt)
		})
	})

	t.Run("D", func(t *testing.T) {
		srcD := &TestBeanAnonymousSrcD{
			TestAnonymousValue: TestAnonymousValue{
				TestAnonymousValueInt: 5,
			},
		}

		t.Run("ParseAnonymousStruct: true", func(t *testing.T) {
			dstD := &TestBeanAnonymousDstD{}
			if err := Copy(srcD).SetConfig(
				&Config{
					ParseAnonymousStruct: true,
				},
			).To(dstD); err != nil {
				t.Fatalf("%#v\n", err)
			}
			assert.Equal(t, srcD.TestAnonymousValue.TestAnonymousValueInt, dstD.TestAnonymousValue.TestAnonymousValueInt)
		})
	})
}

type TestBeanMethodErrorSrc struct {
}

var (
	MethodError = errors.New("args.time can't convert to time.Time")
)

func (src *TestBeanMethodErrorSrc) GenerateMethodResult(m map[string]interface{}) (string, error) {
	t, ok := m["time"].(time.Time)
	if ok {
		return t.Format("2006/01/02"), nil
	} else {
		return "", MethodError
	}
}

type TestBeanMethodErrorDst struct {
	TestDstMethod            string  `deepcopy:"method:GenerateMethodResult"`
	TestDstDefaultMethod     uint    `deepcopy:"default:GenerateDefaultMethodResult"`
	TestDstDefaultNullMethod float64 `deepcopy:"default:1.1"`
}

func (dst *TestBeanMethodErrorDst) GenerateDefaultMethodResult(m map[string]interface{}) (uint, error) {
	return 20, nil
}

func TestDeepCopyMethodErrorStruct(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		nowTime := time.Now()
		src := &TestBeanMethodErrorSrc{}
		dst := &TestBeanMethodErrorDst{}
		args := make(map[string]interface{})
		args["time"] = nowTime

		if err := Copy(src).WithArgs(args).To(dst); err != nil {
			t.Fatalf("%#v\n", err)
		}
		result, err := src.GenerateMethodResult(args)
		if err != nil {
			t.Fatalf("%#v\n", err)
		}

		assert.Equal(t, result, dst.TestDstMethod)
	})

	t.Run("error", func(t *testing.T) {
		nowTime := time.Now()
		src := &TestBeanMethodErrorSrc{}

		t.Run("ErrorArgs", func(t *testing.T) {
			args := make(map[string]interface{})
			args["t"] = nowTime
			dst := &TestBeanMethodErrorDst{}
			if err := Copy(src).WithArgs(args).To(dst); err == nil || err != MethodError {
				t.Fatalf("%#v\n", err)
			}
		})

		t.Run("NullArgs", func(t *testing.T) {
			dst := &TestBeanMethodErrorDst{}
			if err := Copy(src).To(dst); err == nil || err != MethodError {
				t.Fatalf("%#v\n", err)
			}
		})
	})
}

func TestDeepCopyDefaultMethodErrorStruct(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		nowTime := time.Now()
		src := &TestBeanMethodErrorSrc{}
		dst := &TestBeanMethodErrorDst{}
		args := make(map[string]interface{})
		args["time"] = nowTime

		if err := Copy(src).WithArgs(args).To(dst); err != nil {
			t.Fatalf("%#v\n", err)
		}
		result, err := dst.GenerateDefaultMethodResult(args)
		if err != nil {
			t.Fatalf("%#v\n", err)
		}

		assert.Equal(t, result, dst.TestDstDefaultMethod)
		assert.Equal(t, 1.1, dst.TestDstDefaultNullMethod)
	})
}

type TestBeanNullSrc struct {
	TestNullString string
	TestNullBool   bool
	TestNullInt    int
	TestNullTime   time.Time
}

type TestBeanNullDst struct {
	TestNullString null.String
	TestNullBool   null.Bool
	TestNullInt    null.Int
	TestNullTime   null.Time
}

func TestDeepCopyNullStruct(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		nowTime := time.Now()
		src := &TestBeanNullSrc{
			TestNullString: "string",
			TestNullBool:   true,
			TestNullInt:    99,
			TestNullTime:   nowTime,
		}
		dst := &TestBeanNullDst{}

		if err := Copy(src).To(dst); err != nil {
			t.Fatalf("%#v\n", err)
		}
		assert.Equal(t, src.TestNullString, dst.TestNullString.ValueOrZero())
		assert.Equal(t, src.TestNullBool, dst.TestNullBool.ValueOrZero())
		assert.Equal(t, src.TestNullInt, dst.TestNullInt.ValueOrZero())
		assert.Equal(t, src.TestNullTime.String(), dst.TestNullTime.ValueOrZero().String())
	})

	t.Run("empty", func(t *testing.T) {
		src := &TestBeanNullSrc{}
		dst := &TestBeanNullDst{}

		if err := Copy(src).To(dst); err != nil {
			t.Fatalf("%#v\n", err)
		}
		assert.Equal(t, "", dst.TestNullString.ValueOrZero())
		assert.Equal(t, false, dst.TestNullBool.ValueOrZero())
		assert.Equal(t, 0, dst.TestNullInt.ValueOrZero())
		assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", dst.TestNullTime.ValueOrZero().String())
	})

	t.Run("reverse", func(t *testing.T) {
		nowTime := time.Now()
		src := &TestBeanNullSrc{}
		dst := &TestBeanNullDst{
			TestNullString: null.StringFrom("string"),
			TestNullBool:   null.BoolFrom(true),
			TestNullInt:    null.IntFrom(99),
			TestNullTime:   null.TimeFrom(nowTime),
		}

		if err := Copy(dst).To(src); err != nil {
			t.Fatalf("%#v\n", err)
		}
		assert.Equal(t, dst.TestNullString.ValueOrZero(), src.TestNullString)
		assert.Equal(t, dst.TestNullBool.ValueOrZero(), src.TestNullBool)
		assert.Equal(t, dst.TestNullInt.ValueOrZero(), src.TestNullInt)
		assert.Equal(t, dst.TestNullTime.ValueOrZero().String(), src.TestNullTime.String())
	})
}

func BenchmarkDeepCopy(b *testing.B) {
	b.Run("normal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			nowTime := time.Now()

			src := &TestBeanSrc{
				TestInt:       101,
				TestBool:      true,
				TestString:    "this is a test",
				TestInterface: "interface",

				TestSrcMethod: 1,
				TestSrcFrom:   "11",
				TestSrcTo:     1.1,
				TestSkip:      "22",
				TestForce:     &nowTime,
				TestSql: sql.NullInt64{
					Int64: 33,
					Valid: true,
				},
				TestTimeFormatPtr: &nowTime,

				TestStruct: TestStruct{
					TestStructInt: 2,
				},
				TestStructPtr: &TestStruct{
					TestStructInt: 4,
				},
				TestStructPtrToValue: &TestStruct{
					TestStructInt: 7,
				},
				TestStructValueToPtr: TestStruct{
					TestStructInt: 8,
				},

				TestAnonymousValue: TestAnonymousValue{
					TestAnonymousValueInt: 5,
				},
				TestAnonymousValueInt: 99,
				TestAnonymousPtr: &TestAnonymousPtr{
					TestAnonymousPtrInt: 6,
				},
			}
			dst := &TestBeanDst{}

			args := make(map[string]interface{})
			args["time"] = nowTime

			if err := Copy(src).WithArgs(args).To(dst); err != nil {
				b.Fatalf("%#v\n", err)
			}
		}
	})
}

type BenchmarkSrc struct {
	FieldString string
	FieldInt    int
	FieldBool   bool
}

type BenchmarkDst struct {
	FieldString string
	FieldInt    int
	FieldBool   bool
}

func BenchmarkCopy(b *testing.B) {
	src := BenchmarkSrc{
		FieldString: "a",
		FieldInt:    1,
		FieldBool:   true,
	}
	dst := new(BenchmarkDst)
	for i := 0; i < 20000; i++ {
		_ = Copy(&src).To(dst)
	}
}

func TestDeepCopyZeroStruct(t *testing.T) {
	t.Run("timeformat", func(t *testing.T) {
		nowTime := time.Now()

		src := &TestBeanSrc{
			TestTimeFormatPtr: nil,
		}
		dst := &TestBeanDst{}

		args := make(map[string]interface{})
		args["time"] = nowTime

		if err := Copy(src).WithArgs(args).To(dst); err != nil {
			t.Fatalf("%#v\n", err)
		}
	})

	t.Run("force", func(t *testing.T) {
		nowTime := time.Now()

		src := &TestBeanSrc{
			TestForce: nil,
		}
		dst := &TestBeanDst{}

		args := make(map[string]interface{})
		args["time"] = nowTime

		if err := Copy(src).WithArgs(args).To(dst); err != nil {
			t.Fatalf("%#v\n", err)
		}
	})

	t.Run("ptr-value", func(t *testing.T) {
		nowTime := time.Now()

		src := &TestBeanSrc{
			TestStructPtrToValue: nil,
		}
		dst := &TestBeanDst{}

		args := make(map[string]interface{})
		args["time"] = nowTime

		if err := Copy(src).WithArgs(args).To(dst); err != nil {
			t.Fatalf("%#v\n", err)
		}
	})

	t.Run("ptr", func(t *testing.T) {
		nowTime := time.Now()

		src := &TestBeanSrc{
			TestStructPtr: nil,
		}
		dst := &TestBeanDst{}

		args := make(map[string]interface{})
		args["time"] = nowTime

		if err := Copy(src).WithArgs(args).To(dst); err != nil {
			t.Fatalf("%#v\n", err)
		}
	})

	t.Run("anonymous-ptr", func(t *testing.T) {
		nowTime := time.Now()

		src := &TestBeanSrc{
			TestAnonymousPtr: nil,
		}
		dst := &TestBeanDst{}

		args := make(map[string]interface{})
		args["time"] = nowTime

		if err := Copy(src).WithArgs(args).To(dst); err != nil {
			t.Fatalf("%#v\n", err)
		}
	})
}
