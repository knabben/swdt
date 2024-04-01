package setup

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"swdt/apis/config/v1alpha1"
	"testing"
)

var (
	calls      = []string{}
	chocoCheck = fmt.Sprintf("%s --version", CHOCO_PATH)
)

func validateRun(args string) (string, error) {
	calls = append(calls, args)
	return "cmd stdout", nil
}

func TestChocoExist(t *testing.T) {
	calls = []string{}
	expectedCalls := 1

	r := SetupRunner{run: validateRun}
	assert.True(t, r.ChocoExists())
	assert.Len(t, calls, expectedCalls)
	assert.Equal(t, calls[0], chocoCheck)
}

func TestInstallChocoPackages(t *testing.T) {
	calls = []string{}
	expectedCalls := 3
	pkgs := []string{"vim", "grep"}

	r := SetupRunner{run: validateRun}
	config := v1alpha1.AuxiliarySpec{ChocoPackages: &pkgs}
	err := r.InstallChocoPackages(*config.ChocoPackages)
	assert.Nil(t, err)

	assert.Len(t, calls, expectedCalls)
	assert.Equal(t, calls[0], chocoCheck)
	for i := 0; i < expectedCalls-1; i++ {
		assert.Equal(t, calls[i+1], fmt.Sprintf("%s %s %s", CHOCO_PATH, CHOCO_INSTALL, pkgs[i]))
	}
}

func TestEnableRDP(t *testing.T) {
	var defaultTrue = true
	r := SetupRunner{run: validateRun}
	config := v1alpha1.AuxiliarySpec{EnableRDP: &defaultTrue}

	err := r.EnableRDP(*config.EnableRDP)
	assert.Nil(t, err)
}
