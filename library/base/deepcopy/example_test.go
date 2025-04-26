package deepcopy

import (
	"errors"
	"fmt"
	"time"
)

type Src struct {
	Field     string
	FieldTime time.Time
}

func (s *Src) GenerateField(args map[string]interface{}) (int, error) {
	if f, ok := args["field"]; ok {
		return f.(int), nil
	}
	return 0, errors.New("null args")
}

type Dst struct {
	Field string
}

func ExampleDeepCopier_To() {
	src := Src{
		Field: "a",
	}
	dst := new(Dst)
	err := Copy(&src).To(dst)
	if err != nil {
		return
	}
	fmt.Printf("%s\n", dst.Field)

	// OutPut:
	// a
}

func ExampleDeepCopier_From() {
	src := Src{
		Field: "a",
	}
	dst := new(Dst)
	err := Copy(dst).From(&src)
	if err != nil {
		return
	}
	fmt.Printf("%s\n", dst.Field)

	// OutPut:
	// a
}

type DstSkip struct {
	Field string `deepcopy:"skip"`
}

func ExampleTagSkip() {
	src := Src{
		Field: "a",
	}
	dst := new(DstSkip)
	err := Copy(&src).To(dst)
	if err != nil {
		return
	}
	fmt.Printf("%s\n", dst.Field)

	// OutPut:
	//
}

type DstMethod struct {
	Field int `deepcopy:"method:GenerateField"`
}

func ExampleTagMethod() {
	src := Src{
		Field: "a",
	}
	dst := new(DstMethod)
	err := Copy(&src).AddArg("field", 1).To(dst)
	if err != nil {
		return
	}
	fmt.Printf("%d\n", dst.Field)

	dst = new(DstMethod)
	err = Copy(&src).To(dst)
	fmt.Printf("%s\n", err)

	// OutPut:
	// 1
	// null args
}

type DstFrom struct {
	FieldFrom string `deepcopy:"from:Field"`
}

func ExampleTagFrom() {
	src := Src{
		Field: "a",
	}
	dst := new(DstFrom)
	err := Copy(&src).To(dst)
	if err != nil {
		return
	}
	fmt.Printf("%s\n", dst.FieldFrom)

	// OutPut:
	// a
}

type SrcTo struct {
	FieldTo string `deepcopy:"to:Field"`
}

func ExampleTagTo() {
	src := SrcTo{
		FieldTo: "a",
	}
	dst := new(Dst)
	err := Copy(&src).To(dst)
	if err != nil {
		return
	}
	fmt.Printf("%s\n", dst.Field)

	// OutPut:
	// a
}

type DstForce struct {
	Field interface{} `deepcopy:"force"`
}

func ExampleTagForce() {
	src := Src{
		Field: "a",
	}
	dst := new(DstForce)
	err := Copy(&src).To(dst)
	if err != nil {
		return
	}
	fmt.Printf("%s\n", dst.Field)

	// OutPut:
	// a
}

type DstTimeFormat struct {
	FieldTime string `deepcopy:"timeformat:2006/01/02"`
}

func ExampleTagTimeFormat() {
	t, err := time.Parse("2006/01/02 15:04:05", "2019/01/02 11:36:28")
	if err != nil {
		return
	}
	src := Src{
		FieldTime: t,
	}
	dst := new(DstTimeFormat)
	err = Copy(&src).To(dst)
	if err != nil {
		return
	}
	fmt.Printf("%s\n", dst.FieldTime)

	// OutPut:
	// 2019/01/02
}

type DstDefault struct {
	FieldDefaultMethod string `deepcopy:"default:GenerateDefaultString"`
	FieldDefaultValue  string `deepcopy:"default:default value"`
}

func (d *DstDefault) GenerateDefaultString(args map[string]interface{}) string {
	return "default method"
}

func ExampleTagDefault() {
	src := Src{}
	dst := new(DstDefault)
	err := Copy(&src).To(dst)
	if err != nil {
		return
	}
	fmt.Printf("%s\n", dst.FieldDefaultMethod)
	fmt.Printf("%s\n", dst.FieldDefaultValue)

	// OutPut:
	// default method
	// default value
}

type DstAnonymousStruct struct {
	Field string
}

type DstWrapAnonymousStruct struct {
	DstAnonymousStruct
}

func ExampleConfigParseAnonymousStruct() {
	src := Src{
		Field: "a",
	}
	dst := new(DstWrapAnonymousStruct)
	err := Copy(&src).To(dst)
	if err != nil {
		return
	}
	fmt.Printf("%s\n", dst.Field)

	dst = new(DstWrapAnonymousStruct)
	err = Copy(&src).SetConfig(&Config{ParseAnonymousStruct: true}).To(dst)
	if err != nil {
		return
	}
	fmt.Printf("%s\n", dst.Field)

	// OutPut:
	//
	// a
}
