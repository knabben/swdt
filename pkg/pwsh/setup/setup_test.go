package setup

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"swdt/apis/config/v1alpha1"
	"testing"
)

var (
	calls      []string
	chocoCheck = fmt.Sprintf("%s --version", CHOCO_PATH)
)

type Response struct {
	response string
	error    error
}

func TestChocoExist(t *testing.T) {
	responses := []Response{
		{
			response: "",
			error:    nil,
		},
	}
	r := startRunner(&responses)
	assert.True(t, r.ChocoExists())
	assertCalls(t, []string{"choco.exe --version"})
}

func TestInstallChocoPackages(t *testing.T) {
	responses := []Response{
		{
			response: "",
			error:    nil,
		},
		{
			response: "",
			error:    nil,
		},
		{
			response: "",
			error:    nil,
		},
	}
	r := startRunner(&responses)
	config := v1alpha1.AuxiliarySpec{ChocoPackages: &[]string{"vim", "grep"}}
	err := r.InstallChocoPackages(*config.ChocoPackages)
	assert.Nil(t, err)
}

func TestEnableRDP(t *testing.T) {
	responses := []Response{
		{
			response: "",
			error:    nil,
		},
	}
	r := startRunner(&responses)

	var defaultTrue = true
	config := v1alpha1.AuxiliarySpec{EnableRDP: &defaultTrue}
	err := r.EnableRDP(*config.EnableRDP)
	assert.Nil(t, err)
}

func TestInstallContainerdSkip(t *testing.T) {
	responses := []Response{
		{
			response: "Running",
			error:    nil,
		},
		{
			response: "",
			error:    errors.New("error"),
		},
	}

	r := startRunner(&responses)
	config := v1alpha1.ClusterSpec{CalicoVersion: "v3.27"}
	err := r.InstallContainerd(config.CalicoVersion)
	assert.Nil(t, err)
}

func TestInstallContainerdRunning(t *testing.T) {
	responses := []Response{
		{
			response: "",
			error:    errors.New(""),
		},
		{
			response: "",
			error:    nil,
		},
	}

	r := startRunner(&responses)
	err := r.InstallContainerd("v3.27")
	assert.Nil(t, err)
	assertCalls(t, []string{"get-service -name containerd", ".\\Install-Containerd"})
}

func TestInstallKubernetes(t *testing.T) {
	responses := []Response{
		{
			response: "",
			error:    errors.New(""),
		},
		{
			response: "",
			error:    nil,
		},
	}

	r := startRunner(&responses)
	err := r.InstallKubernetes("v1.29.0")
	assert.Nil(t, err)
	assertCalls(t, []string{"get-service -name kubelet", ".\\PrepareNode.ps1 -KubernetesVersion v1.29.0"})
}

func assertCalls(t *testing.T, rcalls []string) {
	assert.Len(t, calls, len(rcalls))
	for idx, call := range rcalls {
		assert.Contains(t, calls[idx], call)
	}
}

func startRunner(responses *[]Response) SetupRunner {
	calls = []string{}
	return SetupRunner{run: func(cmd string) (string, error) {
		calls = append(calls, cmd)
		return popResponse(responses)
	}}
}

func popResponse(responses *[]Response) (string, error) {
	var resp Response
	resp, *responses = (*responses)[0], (*responses)[1:]
	return resp.response, resp.error
}
