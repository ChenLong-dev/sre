package deepcopy

// 配置文件
type Config struct {
	// 严格模式
	// 该模式下，若找不到指定tag，则会返回错误
	StrictMode bool

	// 非零模式
	// 该模式下，针对源为结构体，只拷贝非零值
	NotZeroMode bool

	// 全遍历模式
	// 该模式下，即使源结构体与目标结构体拥有相同类型结构体，仍会遍历，逐个判断
	FullTraversalMode bool

	// 开启可选标签数组
	EnableOptionalTags []string
}
