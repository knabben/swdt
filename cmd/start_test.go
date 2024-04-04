package cmd

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestAlreadyExists(t *testing.T) {
	assert.True(t, alreadyExists(errors.New("already exists with")))
	assert.False(t, alreadyExists(errors.New("error")))
}

func TestStartMinikube(t *testing.T) {
	var called = 0
	fn := func(cmd string, args ...string) (string, error) {
		called += 1
		return "", nil
	}
	err := startMinikube(fn, "v1.29.0")
	assert.Nil(t, err)
	assert.Equal(t, 1, called)
}
