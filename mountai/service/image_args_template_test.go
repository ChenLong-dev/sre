package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestService_renderImageArgsTemplate(t *testing.T) {
	testVariables := map[string]string{
		"SECRET1": "VALUE1",
		"SECRET2": "VALUE!&*()",
	}

	t.Run("replace secret", func(t *testing.T) {
		testContent := "--build-arg KEY=${SECRET1} --build-arg KEY=${SECRET2}"

		imageArg, imageArgWithMask := s.renderImageArgsTemplate(
			context.Background(),
			testContent,
			testVariables,
		)

		assert.Equal(t, imageArg, fmt.Sprintf("--build-arg KEY=\"%s\" --build-arg KEY=\"%s\"",
			testVariables["SECRET1"], testVariables["SECRET2"]))
		assert.Equal(t, imageArgWithMask, fmt.Sprintf("--build-arg KEY=%s --build-arg KEY=%s", MaskSecretStr, MaskSecretStr))
	})

	t.Run("secret not match", func(t *testing.T) {
		testContent := "--build-arg KEY=${secret1} --build-arg KEY=${SECRET2}"

		imageArg, imageArgWithMask := s.renderImageArgsTemplate(
			context.Background(),
			testContent,
			testVariables,
		)

		assert.Equal(t, imageArg, fmt.Sprintf("--build-arg KEY=${secret1} --build-arg KEY=\"%s\"", testVariables["SECRET2"]))
		assert.Equal(t, imageArgWithMask, fmt.Sprintf("--build-arg KEY=${secret1} --build-arg KEY=%s", MaskSecretStr))
	})
}
