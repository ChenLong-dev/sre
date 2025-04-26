package queue

import (
	"gitlab.shanhai.int/sre/library/base/ctime"
	render "gitlab.shanhai.int/sre/library/base/logrender"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"

	"fmt"
	"time"
)

// AMQP客户端
type Queue struct {
	// 配置文件
	config *Config
}

func New(config *Config) *Queue {
	if config == nil {
		panic("queue config is nil")
	}

	if config.Config == nil {
		config.Config = &render.Config{}
	}
	if config.Config.StdoutPattern == "" {
		config.Config.StdoutPattern = defaultPattern
	}
	if config.Config.OutPattern == "" {
		config.Config.OutPattern = defaultPattern
	}
	if config.Config.OutFile == "" {
		config.Config.OutFile = _infoFile
	}
	if config.ConnectTimeout == 0 {
		config.ConnectTimeout = ctime.Duration(DefaultConnectTimeout)
	}
	if config.ReconnectDelay == 0 {
		config.ReconnectDelay = ctime.Duration(DefaultReconnectDelay)
	}
	if config.ReInitDelay == 0 {
		config.ReInitDelay = ctime.Duration(DefaultReInitDelay)
	}

	queue := &Queue{
		config: config,
	}

	return queue
}

func getConnectAddr(cfg *Config) string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d/%s", cfg.UserName, cfg.Password,
		cfg.Endpoint.Address, cfg.Endpoint.Port, cfg.VHost)
}

// 新建会话
func (q *Queue) NewSession(name string) (*Session, error) {
	sc, ok := q.config.Session[name]
	if !ok {
		return nil, errors.New(fmt.Sprintf("session %s haven't config", name))
	}

	if sc.QueueName == "" {
		sc.QueueName = sc.Name
	}
	if sc.ConnectTimeout == 0 {
		sc.ConnectTimeout = q.config.ConnectTimeout
	}
	if sc.ReconnectDelay == 0 {
		sc.ReconnectDelay = q.config.ReconnectDelay
	}
	if sc.ReInitDelay == 0 {
		sc.ReInitDelay = q.config.ReInitDelay
	}

	session := Session{
		id:                 uuid.NewV4().String(),
		config:             sc,
		manager:            NewHookManager(q.config.Config),
		done:               make(chan bool),
		notifyStreamCloses: make([]chan bool, 0),
	}

	// 开启重连协程
	go session.handleReconnect(getConnectAddr(q.config))

	// 等待初始化成功
	connectTimeout := time.NewTicker(time.Duration(sc.ConnectTimeout))
	checkReady := time.NewTicker(connectCheckInterval)
	for {
		if session.IsReady() {
			return &session, nil
		}

		select {
		case <-connectTimeout.C:
			_ = session.Close()
			return nil, ErrNotConnected
		case <-checkReady.C:
			continue
		}
	}
}
