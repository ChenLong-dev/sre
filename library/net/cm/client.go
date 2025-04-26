package cm

// 配置中心客户端
type Client struct {
	// 配置文件
	Config *Config
}

// 新建默认配置中心客户端
func DefaultClient() *Client {
	return NewClient(DefaultConfig())
}

// 新建客户端
func NewClient(config *Config) *Client {
	return &Client{Config: config}
}
