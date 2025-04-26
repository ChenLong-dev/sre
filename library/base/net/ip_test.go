package net

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetInternalIP(t *testing.T) {
	t.Run("local", func(t *testing.T) {
		ip, err := GetInternalIP()
		assert.Nil(t, err)
		fmt.Println(ip)
	})
}
