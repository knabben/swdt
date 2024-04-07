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
