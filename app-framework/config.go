package framework

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/base/ctime"
	"gitlab.shanhai.int/sre/library/base/deepcopy.v2"
	"gitlab.shanhai.int/sre/library/goroutine"
	"gitlab.shanhai.int/sre/library/kafka"
	"gitlab.shanhai.int/sre/library/log"
	"gitlab.shanhai.int/sre/library/net/gin"
	"gitlab.shanhai.int/sre/library/net/sentry"
	"gitlab.shanhai.int/sre/library/net/tracing"
	"gitlab.shanhai.int/sre/library/queue"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
)

const (
	SystemEnvAppID                 = "AMS_APP_ID"
	SystemEnvAppName               = "AMS_APP_NAME"
	SystemEnvProjectName           = "AMS_PROJECT_NAME"
	SystemEnvProjectID             = "AMS_PROJECT_ID"
	SystemEnvAppEnvironment        = "AMS_APP_ENV"
	SystemEnvAmsHost               = "AMS_HOST"
	SystemEnvAmsAuthorizationToken = "AMS_AUTHORIZATION_TOKEN"
)

// RPC 配置
type RPCConfig struct {
	// 服务地址
	Endpoint struct {
		Address string `yaml:"address"`
		Port    int    `yaml:"port"`
	} `yaml:"endpoint"`
	// 超时时间
	Timeout ctime.Duration `yaml:"timeout"`
}

// app 配置
type AppConfig struct {
	// 项目名称
	ProjectName string `yaml:"projectName"`
	// 项目ID
	ProjectID string `yaml:"projectID"`
	// 应用名称
	AppName string `yaml:"appName"`
	// 应用ID
	AppID string `yaml:"appID"`
	// 应用环境
	Env string `yaml:"env"`
}

// 基础配置
type Config struct {
	// app配置文件
	*AppConfig `yaml:"app,inline"`
	// 服务请求的超时时间
	// 服务关闭的超时时间
	ShunDownTimeout ctime.Duration `yaml:"shunDownTimeout"`

	// 健康检查路由名
	HealthCheckRouter string `yaml:"healthCheckRouter"`
	// 数据监控路由名
	MetricsRouter string `yaml:"metricsRouter"`
	// 是否关闭pprof路由
	DisablePProf bool `yaml:"disablePProf"`
	// pprof监听端口号，默认为8089
	PProfPort int `yaml:"pprofPort"`
	// 是否关闭数据统计
	DisableMetrics bool `yaml:"disableMetrics"`
	// metrics监听端口号，默认为8088
	MetricsPort int `yaml:"metricsPort"`
	// 是否关闭链路跟踪
	DisableTracing bool `yaml:"disableTracing"`
	// 是否关闭异常捕获
	DisableCatchPanic bool `yaml:"disableCatchPanic"`
	// 是否关闭Sentry异常捕获
	DisableSentry bool `yaml:"disableSentry"`
	// 是否关闭流量控制
	DisableTrafficShaping bool `yaml:"disableTrafficShaping"`
	// 限流qps
	TrafficShapingQPS float64 `yaml:"trafficShapingQPS"`
	// 限流并发
	TrafficShapingConcurrency float64 `yaml:"trafficShapingConcurrency"`

	// Http服务配置文件
	Gin *gin.Config `yaml:"gin"`
	// RPC服务配置文件
	RPC *RPCConfig `yaml:"rpc"`
	// AMQP消息队列订阅服务配置文件
	AMQP map[string]*queue.Config `yaml:"amqp"`
	// Kafka消息队列订阅服务配置文件
	Kafka map[string]*kafka.Config `yaml:"kafka"`
	// 日志配置文件
	Log *log.Config `yaml:"log"`
	// goroutine配置文件
	Goroutine *goroutine.Config `yaml:"goroutine"`
	// tracing配置文件
	Trace *tracing.Config `yaml:"trace"`
	// sentry配置文件
	Sentry *sentry.Config `yaml:"sentry"`
}

// 从配置文件中解码到相应的配置结构体
//
//	configPath为配置文件的所属位置，例如 ./config/config.yaml
//	config为需要解析到的结构体指针
func DecodeConfig(configPath string, config interface{}) {
	if configPath == "" {
		panic("config path is empty")
	}

	if config == nil {
		panic("config is nil")
	}
	if reflect.ValueOf(config).IsNil() {
		panic("config value is nil")
	}
	if reflect.TypeOf(config).Kind() != reflect.Ptr {
		panic("config value is not ptr")
	}

	err := DecodeConfigFromLocal(configPath, config)
	if err != nil {
		panic(err)
	}

	baseConf := &Config{
		Sentry: &sentry.Config{},
	}

	err = deepcopy.Copy(config).SetConfig(&deepcopy.Config{
		NotZeroMode: true,
	}).To(baseConf)

	if err != nil {
		panic(err)
	}

	// 从环境变量获取app配置
	decodeConfigFromEnv(baseConf)

	// 调ams api获取app配置
	err = decodeConfigFromApi(baseConf)
	if err != nil {
		log.Error("decode config from api error:%s", err.Error())
	}

	if baseConf.Sentry.Environment == "" {
		baseConf.Sentry.Environment = baseConf.Env
	}

	err = deepcopy.Copy(baseConf).SetConfig(&deepcopy.Config{
		NotZeroMode: true,
	}).To(config)

	if err != nil {
		panic(err)
	}
}

// 从配置文件中解码到相应的配置结构体
//
//	configPath为配置文件的所属位置，例如 ./config/config.yaml
//	config为需要解析到的结构体指针
func DecodeConfigFromLocal(configPath string, config interface{}) error {
	configFile, err := os.Open(configPath)
	if err != nil {
		return err
	}
	defer configFile.Close()

	configData, err := ioutil.ReadAll(configFile)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(configData, config)
	if err != nil {
		return err
	}

	return nil
}

// 从环境变量里读取配置
func decodeConfigFromEnv(config *Config) {
	config.AppID = initConfFromEnv(config.AppID, SystemEnvAppID)
	config.AppName = initConfFromEnv(config.AppName, SystemEnvAppName)
	config.ProjectName = initConfFromEnv(config.ProjectName, SystemEnvProjectName)
	config.ProjectID = initConfFromEnv(config.ProjectID, SystemEnvProjectID)
	config.Env = initConfFromEnv(config.Env, SystemEnvAppEnvironment)
}

// 调ams接口获取app配置响应
type amsAppDetailResp struct {
	AmsAppDetail *AmsAppDetail `json:"data"`
	ErrCode      int           `json:"errcode"`
	ErrMsg       string        `json:"errmsg"`
}

// app基础信息
type AmsAppDetail struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ProjectID   string `json:"project_id"`
	ProjectName string `json:"project_name"`
	SentryDsn   string `json:"sentry_project_public_dsn"`
}

// 调接口获取配置
func decodeConfigFromApi(config *Config) error {
	// 获取不到host时，则不调用接口
	amsHost := os.Getenv(SystemEnvAmsHost)
	if amsHost == "" {
		return nil
	}
	token := os.Getenv(SystemEnvAmsAuthorizationToken)
	if token == "" {
		return nil
	}

	if config.AppID == "" {
		return errors.New("appID can not be empty")
	}
	if config.Env == "" {
		return errors.New("env must be set")
	}

	client := &http.Client{}
	url := fmt.Sprintf("%s/api/v1/apps/%s?%s=%s&%s=%s", amsHost, config.AppID, "env_name", config.Env, "data_type", "general")
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("get ams error: status_code: %d", resp.StatusCode)
	}

	defer resp.Body.Close()
	allData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	getConfigResp := new(amsAppDetailResp)
	err = json.Unmarshal(allData, &getConfigResp)
	if err != nil {
		return err
	}
	if getConfigResp.ErrCode != 0 {
		return fmt.Errorf("get ams error: errcode:%d errmsg:%s", getConfigResp.ErrCode, getConfigResp.ErrMsg)
	}

	if config.Sentry.DSN == "" {
		config.Sentry.DSN = getConfigResp.AmsAppDetail.SentryDsn
	}
	if config.ProjectName == "" {
		config.ProjectName = getConfigResp.AmsAppDetail.ProjectName
	}
	if config.ProjectID == "" {
		config.ProjectID = getConfigResp.AmsAppDetail.ProjectID
	}
	if config.AppName == "" {
		config.AppName = getConfigResp.AmsAppDetail.Name
	}
	if config.AppID == "" {
		config.AppID = getConfigResp.AmsAppDetail.ID
	}

	return nil
}

// 从环境变量更新app配置信息
func initConfFromEnv(appConf string, envKey string) string {
	if appConf == "" {
		if value := os.Getenv(envKey); value != "" {
			appConf = value
		}
	}

	return appConf
}
