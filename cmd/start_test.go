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
	var called int = 0
	fn := func(cmd string, args ...string) error {
		called += 1
		return nil
	}
	err := startMinikube(fn)
	assert.Nil(t, err)
	assert.Equal(t, 1, called)
}

func TestStartWindowsVMValidateDisk(t *testing.T) {
	diskPath = ""
	assert.Error(t, validateFlags())

	diskPath = "some-path"
	assert.Nil(t, validateFlags())
}
