package cm

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestClient_GetOriginFile(t *testing.T) {
	t.Run("stg", func(t *testing.T) {
		err := os.Setenv("env", "stg")
		assert.Nil(t, err)

		_, err = DefaultClient().GetOriginFile(
			"projects/app-framework/stg/config.yaml",
			"b413878fbdf5d2d658b386f9b4de109ac5dc2f40",
		)
		assert.Nil(t, err)
	})

	t.Run("prd", func(t *testing.T) {
		err := os.Setenv("env", "stg")
		assert.Nil(t, err)

		_, err = DefaultClient().GetOriginFile(
			"projects/app-framework/stg/config.yaml",
			"b413878fbdf5d2d658b386f9b4de109ac5dc2f40",
		)
		assert.Nil(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		err := os.Setenv("env", "stg")
		assert.Nil(t, err)

		_, err = DefaultClient().GetOriginFile(
			"projects/app_framework/stg/abcdefg.yaml",
			"b413878fbdf5d2d658b386f9b4de109ac5dc2f40",
		)
		assert.NotNil(t, err)
	})
}
