package test

import (
	"github.com/stretchr/testify/assert"

	"net"
	"testing"
	"time"
)

func TestRunPProfInBackground(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		RunPProfInBackground("")
		_, err := net.DialTimeout("tcp", DefaultUnitPProfAddr, time.Second*3)
		assert.NoError(t, err)
	})
}
