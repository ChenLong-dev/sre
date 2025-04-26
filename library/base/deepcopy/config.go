package deepcopy

// 配置文件
// Deprecated
// 推荐使用v2版本
type Config struct {
	// 是否解析出匿名结构体中字段，默认为false
	// 当该配置为true时：
	// 若目标结构体中存在与匿名结构体中同名字段，则只会赋值外部同名字段，匿名结构体内部字段不会被赋值
	// 例：
	// 	type A struct {
	// 			S int
	// 	}
	// 	type B struct {
	//			A
	// 	}
	// 	type C struct {
	// 			S	int
	//			A
	// 	}
	//	B.A.S=1
	// 	Copy(B).To(C) // C.A.S=0, C.S=1
	// 若源结构体中存在与匿名结构体中同名字段，则只会取外部同名字段，而不会取匿名结构体内部字段
	// 例：
	// 	type A struct {
	// 			S int
	// 	}
	// 	type B struct {
	//			A
	//			S 	int
	// 	}
	// 	type C struct {
	//			A
	// 	}
	//	B.A.S=1
	//	B.S=2
	// 	Copy(B).To(C) // C.A.S=2
	ParseAnonymousStruct bool
}
