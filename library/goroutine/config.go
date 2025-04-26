package goroutine

import render "gitlab.shanhai.int/sre/library/base/logrender"

var (
	conf *Config
)

// 默认初始化
func init() {
	conf = &Config{
		Config: &render.Config{
			Stdout: false,
			OutDir: "",
		},
	}
}

type Config struct {
	// 日志配置
	*render.Config `yaml:",inline"`
}

// 外部初始化
func Init(c *Config) {
	if c != nil {
		conf = c
	}

	if conf.Config == nil {
		conf.Config = &render.Config{}
	}
	if conf.Config.StdoutPattern == "" {
		conf.Config.StdoutPattern = defaultPattern
	}
	if conf.Config.OutPattern == "" {
		conf.Config.OutPattern = defaultPattern
	}
	if conf.Config.OutFile == "" {
		conf.Config.OutFile = _infoFile
	}

	_manager = NewHookManager(conf.Config)
}

// 关闭
func Close() error {
	_manager.Close()
	return nil
}
