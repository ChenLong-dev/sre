package queue

import (
	"gitlab.shanhai.int/sre/library/base/ctime"
	render "gitlab.shanhai.int/sre/library/base/logrender"
	"gitlab.shanhai.int/sre/library/internal/test"

	toxiproxy "github.com/Shopify/toxiproxy/client"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"

	"os"
	"path/filepath"
	"testing"
	"time"
)

var (
	globalConfig Config
	toxiProxy    *toxiproxy.Proxy
)

func TestMain(m *testing.M) {
	test.RunPProfInBackground("")

	var unitConfig *test.UnitConfig
	toxiProxy, unitConfig = test.InitUnitProxy(
		filepath.Join("..", test.UnitConfigPath),
		"queue",
	)

	globalConfig = Config{
		Endpoint: &EndpointConfig{
			Address: unitConfig.Queue.Address,
			Port:    unitConfig.Queue.Port,
		},
		UserName:       unitConfig.Queue.UserName,
		Password:       unitConfig.Queue.Password,
		VHost:          unitConfig.Queue.VHost,
		ConnectTimeout: ctime.Duration(time.Second * 10),
		ReconnectDelay: ctime.Duration(time.Second * 5),
		ReInitDelay:    ctime.Duration(time.Second * 5),
		Session: map[string]*SessionConfig{
			"unit": {
				QueueName:      "unittest",
				ExchangeName:   "boot",
				RoutingKey:     "",
				Durable:        true,
				AutoDelete:     false,
				Exclusive:      false,
				NoWait:         false,
				ConnectTimeout: ctime.Duration(time.Second * 15),
				ReconnectDelay: ctime.Duration(time.Second * 3),
				ReInitDelay:    ctime.Duration(time.Second * 3),
			},
		},
		Config: &render.Config{
			Stdout: true,
		},
	}

	code := m.Run()

	test.CloseUnitProxy()

	os.Exit(code)
}

var testNew = []test.SyncUnitCase{
	{
		Name: "empty config",
		Func: func(t *testing.T) {
			assert.Panics(t, func() {
				New(nil)
			})
		},
	},
	{
		Name: "default config",
		Func: func(t *testing.T) {
			q := New(&Config{})
			assert.NotNil(t, q)
			assert.NotNil(t, q.config)

			c := q.config
			assert.NotNil(t, c.Config)
			assert.Equal(t, defaultPattern, c.StdoutPattern)
			assert.Equal(t, defaultPattern, c.OutPattern)
			assert.Equal(t, _infoFile, c.OutFile)

			assert.Equal(t, DefaultConnectTimeout, time.Duration(c.ConnectTimeout))
			assert.Equal(t, DefaultReconnectDelay, time.Duration(c.ReconnectDelay))
			assert.Equal(t, DefaultReInitDelay, time.Duration(c.ReInitDelay))
		},
	},
}

var testQueueNewSession = []test.SyncUnitCase{
	{
		Name: "invalid session",
		Func: func(t *testing.T) {
			q := New(&globalConfig)
			s, err := q.NewSession("unknown")
			assert.Nil(t, s)
			if assert.Error(t, err) {
				assert.Equal(t, "session unknown haven't config", err.Error())
			}
		},
	},
	{
		Name: "default config",
		Func: func(t *testing.T) {
			globalConfig.Session["default"] = &SessionConfig{
				Name:         "unittest",
				ExchangeName: "boot",
				RoutingKey:   "",
				Durable:      true,
			}
			q := New(&globalConfig)
			s, err := q.NewSession("default")
			assert.NoError(t, err)
			assert.NotNil(t, s)

			qc := q.config
			sc := s.config
			assert.Equal(t, sc.Name, sc.QueueName)
			assert.Equal(t, qc.ConnectTimeout, sc.ConnectTimeout)
			assert.Equal(t, qc.ReconnectDelay, sc.ReconnectDelay)
			assert.Equal(t, qc.ReInitDelay, sc.ReInitDelay)
		},
	},
	{
		Name: "connect timeout",
		Func: func(t *testing.T) {
			_ = toxiProxy.Disable()
			defer func() {
				_ = toxiProxy.Enable()
			}()

			q := New(&globalConfig)
			_, err := q.NewSession("unit")
			assert.EqualError(t, err, ErrNotConnected.Error())
		},
	},
	{
		Name: "reconnect",
		Func: func(t *testing.T) {
			q := New(&globalConfig)
			s, err := q.NewSession("unit")
			assert.NoError(t, err)
			assert.NotNil(t, s)
			assert.Equal(t, true, s.IsReady())

			// 断网
			_ = toxiProxy.Disable()
			time.Sleep(time.Second)
			assert.Equal(t, false, s.IsReady())

			// 恢复网络
			_ = toxiProxy.Enable()
			time.Sleep(time.Second * 3)
			assert.Equal(t, true, s.IsReady())
		},
	},
	{
		Name: "re init before close",
		Func: func(t *testing.T) {
			globalConfig.Session["reinit"] = &SessionConfig{
				QueueName:    "unittest",
				ExchangeName: "boot",
				RoutingKey:   "",
				Durable:      true,
				ReInitDelay:  ctime.Duration(time.Second * 15),
			}
			q := New(&globalConfig)
			s, err := q.NewSession("reinit")
			assert.NoError(t, err)

			// 断网
			_ = toxiProxy.Disable()
			time.Sleep(time.Second)
			assert.Equal(t, false, s.IsReady())

			// 恢复网络
			_ = toxiProxy.Enable()
			time.Sleep(time.Second * 3)
			assert.Equal(t, false, s.IsReady())
		},
	},
	{
		Name: "declare error",
		Func: func(t *testing.T) {
			globalConfig.Session["DeclareError"] = &SessionConfig{
				QueueName:    "unittest",
				ExchangeName: "boot",
				RoutingKey:   "",
				Durable:      true,
				Args: amqp.Table{
					"some": uint(1),
				},
				ReInitDelay: ctime.Duration(time.Second),
			}
			q := New(&globalConfig)
			_, err := q.NewSession("DeclareError")
			assert.Equal(t, err, ErrNotConnected)
		},
	},
}

func TestNew(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		test.RunSyncCases(t, testNew...)
	})

	t.Run("Queue_NewSession", func(t *testing.T) {
		test.RunSyncCases(t, testQueueNewSession...)
	})

	t.Run("Session_CloseStream", func(t *testing.T) {
		test.RunSyncCases(t, testSessionCloseStream...)
	})

	t.Run("Session_CloseChannel", func(t *testing.T) {
		test.RunSyncCases(t, testSessionCloseChannel...)
	})

	t.Run("Session_Close", func(t *testing.T) {
		test.RunSyncCases(t, testSessionClose...)
	})

	t.Run("Session_ConnectionDo", func(t *testing.T) {
		test.RunSyncCases(t, testSessionConnectionDo...)
	})

	t.Run("Session_ChannelDo", func(t *testing.T) {
		test.RunSyncCases(t, testSessionChannelDo...)
	})

	t.Run("Session_SingleConsumeStream", func(t *testing.T) {
		test.RunSyncCases(t, testSessionSingleConsumeStream...)
	})

	t.Run("Session_NoAutoAckConsumeStream", func(t *testing.T) {
		test.RunSyncCases(t, testSessionNoAutoAckConsumeStream...)
	})

	t.Run("Session_Stream", func(t *testing.T) {
		test.RunSyncCases(t, testSessionStream...)
	})

	t.Run("Session_Push", func(t *testing.T) {
		test.RunSyncCases(t, testSessionPush...)
	})

	t.Run("Session_UnsafePush", func(t *testing.T) {
		test.RunSyncCases(t, testSessionUnsafePush...)
	})
}
