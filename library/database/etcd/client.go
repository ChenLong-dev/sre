package etcd

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"strings"
	"time"

	render "gitlab.shanhai.int/sre/library/base/logrender"
	"gitlab.shanhai.int/sre/library/net/cm"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/pkg/transport"
)

// 新建客户端
func NewClient(c *Config) (db *DB) {
	if c == nil {
		panic("etcd config is nil")
	}
	if len(c.Endpoints) == 0 {
		panic("endpoints is empty")
	}
	// 组装端点
	points := make([]string, 0)
	for _, point := range c.Endpoints {
		points = append(points, fmt.Sprintf("%s:%d", point.Address, point.Port))
	}

	if c.Config == nil {
		c.Config = &render.Config{}
	}
	if c.Config.StdoutPattern == "" {
		c.Config.StdoutPattern = defaultPattern
	}
	if c.Config.OutPattern == "" {
		c.Config.OutPattern = defaultPattern
	}
	if c.Config.OutFile == "" {
		c.Config.OutFile = _infoFile
	}

	clientConfig := clientv3.Config{
		Endpoints:   points,
		DialTimeout: time.Duration(c.DialTimeout),
		Username:    c.UserName,
		Password:    c.Password,
	}
	if c.Tls != nil && c.Tls.Enable {
		tlsConfig, err := getTlsConfig(c.Tls)
		if err != nil {
			panic(err)
		}
		clientConfig.TLS = tlsConfig
	}

	client, err := clientv3.New(clientConfig)
	if err != nil {
		panic(err)
	}

	db = &DB{
		client:   client,
		kvClient: clientv3.NewKV(client),
		storeMap: make(map[string]*StoreData),
		config:   c,
		manager:  NewHookManager(c.Config),
		done:     make(chan bool),
	}

	for _, p := range db.config.Preload {
		_, err := db.LoadPrefixData(context.Background(), p.Prefix, p.ValueFilter, p.EnableWatch)
		if err != nil {
			panic(err)
		}
	}

	return db
}

// 获取tls配置
func getTlsConfig(cfg *TlsConfig) (*tls.Config, error) {
	certFile, err := unmarshalTlsFilePath(cfg.CertFilePath)
	if err != nil {
		return nil, err
	}
	keyFile, err := unmarshalTlsFilePath(cfg.KeyFilePath)
	if err != nil {
		return nil, err
	}
	caFile, err := unmarshalTlsFilePath(cfg.TrustedCAFilePath)
	if err != nil {
		return nil, err
	}

	tlsConfig, err := transport.TLSInfo{
		CertFile:      certFile,
		KeyFile:       keyFile,
		TrustedCAFile: caFile,
	}.ClientConfig()
	if err != nil {
		return nil, err
	}
	return tlsConfig, nil
}

// 解析tls文件路径
// 若为本地则直接使用，否则从远程下载
func unmarshalTlsFilePath(path string) (string, error) {
	if !strings.HasPrefix(path, "private/") {
		return path, nil
	}
	data, err := cm.DefaultClient().GetOriginFile(path, "")
	if err != nil {
		return "", err
	}

	pathList := strings.Split(path, "/")
	fileName := fmt.Sprintf("./%s", pathList[len(pathList)-1])
	file, err := os.Create(fileName)
	if err != nil {
		return "", err
	}
	defer file.Close()
	_, err = file.Write(data)
	if err != nil {
		return "", err
	}
	return fileName, nil
}
