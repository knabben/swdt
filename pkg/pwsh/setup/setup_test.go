package setup

import (
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"swdt/apis/config/v1alpha1"
	"swdt/pkg/executors/exec"
	"swdt/pkg/executors/iface"
	"swdt/pkg/executors/tests"
	"testing"
)

var (
	calls []string
	port  = 2222
)

// startServer starts a fake SSH server
func StartServer(port int, expected *[]tests.Response) *v1alpha1.SSHSpec {
	hostname := tests.GetHostname(port)
	tests.NewServer(hostname, expected)
	return &v1alpha1.SSHSpec{
		Hostname: hostname,
		Username: tests.Username,
		Password: tests.FakePassword,
	}
}

type LocalExec struct {
	stdout *chan string
}

func (l LocalExec) Run(args string, stdout *chan string) error {
	return nil
}

func (l LocalExec) Stdout(std *chan string) {
	l.stdout = std
}

func (l LocalExec) Stderr(std *chan string) {
	panic("implement me")
}

func NewLocalExecutor() iface.Executor {
	return &LocalExec{}
}

func startRunner(responses *[]tests.Response) (*Runner, error) {
	calls = []string{}
	credentials := StartServer(port, responses)
	var sshExec = exec.NewSSHExecutor(credentials) // Start SSH Connection
	if err := sshExec.Connect(); err != nil {
		return nil, err
	}
	return &Runner{remote: sshExec, local: NewLocalExecutor()}, nil
}

func TestChocoExist(t *testing.T) {
	responses := &[]tests.Response{
		{
			Response: "v1.0",
			Error:    nil,
		},
	}
	port += 1
	r, err := startRunner(responses)
	assert.Nil(t, err)
	assert.True(t, r.ChocoExists())
}

func TestInstallChocoPackages(t *testing.T) {
	responses := &[]tests.Response{
		{
			Response: "",
			Error:    nil,
		},
		{
			Response: "",
			Error:    nil,
		},
		{
			Response: "",
			Error:    nil,
		},
	}
	port += 1
	r, err := startRunner(responses)
	assert.Nil(t, err)
	config := v1alpha1.AuxiliarySpec{ChocoPackages: &[]string{"vim", "grep"}}
	assert.Nil(t, r.InstallChocoPackages(*config.ChocoPackages))
}

func TestEnableRDP(t *testing.T) {
	var defaultTrue = true
	responses := &[]tests.Response{
		{
			Response: "",
			Error:    nil,
		},
	}
	port += 1
	r, err := startRunner(responses)
	assert.Nil(t, err)
	config := v1alpha1.AuxiliarySpec{EnableRDP: &defaultTrue}
	assert.Nil(t, r.EnableRDP(*config.EnableRDP))
}

func TestInstallContainerdSkip(t *testing.T) {
	responses := &[]tests.Response{
		{
			Response: "Running",
			Error:    nil,
		},
		{
			Response: "",
			Error:    errors.New("error"),
		},
	}
	port += 1
	r, err := startRunner(responses)
	assert.Nil(t, err)
	config := v1alpha1.ClusterSpec{CalicoVersion: "v3.27"}
	err = r.InstallContainerd(config.CalicoVersion)
	assert.Nil(t, err)
}

func TestInstallContainerdRunning(t *testing.T) {
	responses := &[]tests.Response{
		{
			Response: "",
			Error:    errors.New(""),
		},
		{
			Response: "",
			Error:    nil,
		},
	}

	port += 1
	r, err := startRunner(responses)
	assert.Nil(t, err)
	err = r.InstallContainerd("v3.27")
	assert.Nil(t, err)
}

func TestInstallKubernetes(t *testing.T) {
	responses := &[]tests.Response{
		{
			Response: "",
			Error:    errors.New(""),
			Cmd:      "get-service -name kubelet",
		},
		{
			Response: "",
			Error:    nil,
			Cmd:      ".\\PrepareNode.ps1 -KubernetesVersion v1.29.0",
		},
	}
	port += 1
	r, err := startRunner(responses)
	assert.Nil(t, err)

	err = r.InstallKubernetes("v1.29.0")
	assert.Nil(t, err)
}

func TestJoinNodeRunner(t *testing.T) {
	responses := &[]tests.Response{
		{
			Response: "Stopped",
			Error:    nil,
			Cmd:      "get-service -name kubelet",
		},
		{
			Response: "",
			Error:    nil,
			Cmd:      "mkdir",
		},
		{
			Response: "",
			Error:    nil,
			Cmd:      "Add-content",
		},
		{
			Response: "",
			Error:    nil,
			Cmd:      "$env:Path",
		},
	}
	port += 1
	r, err := startRunner(responses)
	assert.Nil(t, err)
	err = r.JoinNode("v1.29.0", "192.168.0.1")
	assert.Nil(t, err)
}
