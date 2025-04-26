package render

type Config struct {
	// 控制台日志是否输出
	Stdout bool `yaml:"stdout"`
	// 控制台日志渲染模版
	StdoutPattern string `yaml:"stdoutPattern"`
	// 日志文件输出目录，空为不输出
	OutDir string `yaml:"outDir"`
	// 日志文件名
	OutFile string `yaml:"outFile"`
	// 日志文件渲染模版
	OutPattern string `yaml:"outPattern"`
	// 日志文件缓冲区大小
	FileBufferSize int64 `yaml:"fileBufferSize"`
	// 同时存在最大日志文件数
	MaxLogFile int `yaml:"maxLogFile"`
	// 日志文件切割大小
	RotateSize int64 `yaml:"rotateSize"`
}
