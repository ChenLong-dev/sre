package test

import (
	"github.com/stretchr/testify/assert"

	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestInitUnitProxy(t *testing.T) {
	configPath := filepath.Join("..", "..", UnitConfigPath)

	cases := []SyncUnitCase{
		{
			"queue",
			func(t *testing.T) {
				proxy, cfg := InitUnitProxy(configPath, "queue")
				assert.NotNil(t, proxy)
				assert.NotNil(t, cfg)

				_, err := net.DialTimeout("tcp",
					net.JoinHostPort(UnitProxyHost, strconv.Itoa(UnitProxyPort)), time.Second*3)
				assert.NoError(t, err)

				CloseUnitProxy()
			},
		},
		{
			"multi",
			func(t *testing.T) {
				proxy, cfg := InitUnitProxy(configPath, "queue")
				assert.NotNil(t, proxy)
				assert.NotNil(t, cfg)

				assert.NotPanics(t, func() {
					proxy, cfg = InitUnitProxy(configPath, "queue")
					assert.NotNil(t, proxy)
					assert.NotNil(t, cfg)

					_, err := net.DialTimeout("tcp",
						net.JoinHostPort(UnitProxyHost, strconv.Itoa(UnitProxyPort)), time.Second*3)
					assert.NoError(t, err)
				})

				CloseUnitProxy()
			},
		},
		{
			"no server",
			func(t *testing.T) {
				// 抢占proxy端口
				var server *http.Server
				address := net.JoinHostPort(UnitProxyHost, strconv.Itoa(UnitProxyPort))
				defer func() {
					if server != nil {
						_ = server.Close()
					}
				}()
				go func() {
					server = &http.Server{
						Addr: address,
					}
					_ = server.ListenAndServe()
				}()

				time.Sleep(InitClientDelay)

				assert.Panics(t, func() {
					listenUnitProxyServer(address)
				})
				assert.Panics(t, func() {
					_, _ = InitUnitProxy(configPath, "unknown")
				})

				CloseUnitProxy()
			},
		},
		{
			"invalid proxy",
			func(t *testing.T) {
				assert.Panics(t, func() {
					_, _ = InitUnitProxy(configPath, "unknown")
				})

				CloseUnitProxy()
			},
		},
	}

	RunSyncCases(t, cases...)
}
