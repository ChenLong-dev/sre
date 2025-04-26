package agollo

import (
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	render "gitlab.shanhai.int/sre/library/base/logrender"
)

func TestClient_newClient(t *testing.T) {
	t.Skip("Should have apollo environment")

	client := NewClient(&Config{
		AppID:             "your_appID",
		Cluster:           "dev",
		ServerHost:        "http://localhost:8080",
		PreloadNamespaces: []string{"application", "namespace1", "namespace2"},
		NotDaemon:         false,
		Config: &render.Config{
			Stdout:        true,
			StdoutPattern: "[%T] [%t] %S , appID: %A , cluster: %C , serveHost: %I , namespace: %N , Extra: %E , changes: %c",
		},
	})
	t.Run("normal", func(t *testing.T) {
		all, err := client.GetNamespace("application")
		assert.Nil(t, err)
		res, err := all.Load("key")
		assert.Nil(t, err)
		assert.Equal(t, "value", res)
	})

	t.Run("together", func(t *testing.T) {
		res, err := client.GetNewValue("application", "key")
		assert.Nil(t, err)
		assert.Equal(t, "value", res)
	})
}

func TestClient_Close(t *testing.T) {
	t.Skip("Should have apollo environment")

	t.Run("daemon_watch", func(t *testing.T) {
		preNum := runtime.NumGoroutine()
		client := NewClient(&Config{
			AppID:             "your_appID",
			Cluster:           "dev",
			ServerHost:        "http://localhost:8080",
			PreloadNamespaces: []string{"application", "namespace1", "namespace2"},
			NotDaemon:         false,
			Config: &render.Config{
				Stdout:        true,
				StdoutPattern: "[%T] [%t] %S , appID: %A , cluster: %C , serveHost: %I , namespace: %N , Extra: %E , changes: %c",
			},
		})

		time.Sleep(time.Second * 3)
		_, err := client.GetNewValue("application", "key")
		assert.Nil(t, err)
		all, err := client.GetNamespace("application")
		assert.Nil(t, err)
		_, err = all.Load("key")
		assert.Nil(t, err)

		time.Sleep(time.Second * 3)
		client.Close()

		time.Sleep(time.Second * 120)

		assert.Equal(t, preNum, runtime.NumGoroutine())
	})

	t.Run("simple_watch", func(t *testing.T) {
		preNum := runtime.NumGoroutine()
		client := NewClient(&Config{
			AppID:             "your_appID",
			Cluster:           "dev",
			ServerHost:        "http://localhost:8080",
			PreloadNamespaces: []string{"application", "namespace1", "namespace2"},
			NotDaemon:         true,
			Config: &render.Config{
				Stdout:        true,
				StdoutPattern: "[%T] [%t] %S , appID: %A , cluster: %C , serveHost: %I , namespace: %N , Extra: %E , changes: %c",
			},
		})

		resp, err := client.Watch([]string{"application"})
		assert.Nil(t, err)

		go func() {
			for {
				select {
				case r, ok := <-resp:
					if ok {
						if r.Type == EventUpdate {
							// 判断namespace是否与监听的namespace一致
							if r.Namespace == "application" {

							}
						} else if r.Type == EventError {
							assert.Nil(t, r.ErrInfo.Err)
						}
					}
				case <-client.Done():
					return
				}
			}
		}()

		time.Sleep(time.Second * 30)
		client.Close()
		time.Sleep(time.Second * 120)

		assert.Equal(t, preNum, runtime.NumGoroutine())
	})

	t.Run("close_when_resp_not_consume", func(t *testing.T) {
		preNum := runtime.NumGoroutine()
		client := NewClient(&Config{
			AppID:             "your_appID",
			Cluster:           "dev",
			ServerHost:        "http://localhost:8080",
			PreloadNamespaces: []string{"application", "namespace1", "namespace2"},
			NotDaemon:         true,
			Config: &render.Config{
				Stdout:        true,
				StdoutPattern: "[%T] [%t] %S , appID: %A , cluster: %C , serveHost: %I , namespace: %N , Extra: %E , changes: %c",
			},
		})
		_, err := client.Watch([]string{"application"})
		assert.Nil(t, err)

		time.Sleep(time.Second * 30)
		client.Close()
		time.Sleep(time.Second * 120)

		assert.Equal(t, preNum, runtime.NumGoroutine())
	})
}

func TestClient_watch(t *testing.T) {
	t.Skip("Should have apollo environment")

	client := NewClient(&Config{
		AppID:             "your_appID",
		Cluster:           "dev",
		ServerHost:        "http://localhost:8080",
		PreloadNamespaces: []string{"application", "namespace1", "namespace2"},
		NotDaemon:         true,
		Config: &render.Config{
			Stdout:        true,
			StdoutPattern: "[%T] [%t] %S , appID: %A , cluster: %C , serveHost: %I , namespace: %N , Extra: %E , changes: %c",
		},
	})
	defer client.Close()

	res := client.Get("application")
	expected := "value"
	assert.Equal(t, expected, res["key"])

	resp, err := client.Watch([]string{"application"})
	assert.Nil(t, err)

	expected = "value"
	go func() {
		for {
			select {
			case r, ok := <-resp:
				if ok {
					if r.Type == EventUpdate {
						if r.Namespace == "application" {
							conf := client.Get(r.Namespace)
							assert.Equal(t, expected, conf)
						}
					} else if r.Type == EventError {
						assert.Nil(t, r.ErrInfo.Err)
					}
				}
			case <-client.Done():
				return
			}
		}
	}()
}

func TestClient_get(t *testing.T) {
	t.Skip("Should have apollo environment")

	client := NewClient(&Config{
		AppID:      "your_appID",
		Cluster:    "dev",
		ServerHost: "http://localhost:8080",
		Config: &render.Config{
			Stdout:        true,
			StdoutPattern: "[%T] [%t] %S , appID: %A , cluster: %C , serveHost: %I , namespace: %N , Extra: %E , changes: %c",
		},
	})
	expected := client.Get("application")
	actual := "value"
	assert.Equal(t, actual, expected["key"])
}
