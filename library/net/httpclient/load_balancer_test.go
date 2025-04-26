package httpclient

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAppNameAndEnvFromK8sHost(t *testing.T) {
	t.Run("us", func(t *testing.T) {
		u, err := url.Parse("http://user-system.prd.svc.cluster.local/u2/api/v4/user/register")
		assert.Nil(t, err)

		name, env, err := ParseNameAndEnvFromK8sHost(u.Host)
		assert.Nil(t, err)
		assert.Equal(t, "user-system", name)
		assert.Equal(t, "prd", env)
	})

	t.Run("dev", func(t *testing.T) {
		u, err := url.Parse("http://dev.user-system.prd.svc.cluster.local/u2/api/v4/user/register")
		assert.Nil(t, err)

		name, env, err := ParseNameAndEnvFromK8sHost(u.Host)
		assert.Nil(t, err)
		assert.Equal(t, "user-system", name)
		assert.Equal(t, "prd", env)
	})

	t.Run("u2", func(t *testing.T) {
		u, err := url.Parse("http://us-inner.qingting-hz.com/u2/api/v4/user/register")
		assert.Nil(t, err)

		_, _, err = ParseNameAndEnvFromK8sHost(u.Host)
		assert.NotNil(t, err)
	})
}
