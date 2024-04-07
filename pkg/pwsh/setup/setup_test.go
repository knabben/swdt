package setup

import (
	"github.com/stretchr/testify/assert"
	"swdt/apis/config/v1alpha1"
	"swdt/pkg/executors/exec"
	"swdt/pkg/executors/tests"
	"testing"
)

var (
	calls []string
	port  = 2222
)

func assertCalls(t *testing.T, rcalls []string) {
	assert.Len(t, calls, len(rcalls))
	for idx, call := range rcalls {
		assert.Contains(t, calls[idx], call)
	}
}

func extractCmds(responses *[]tests.Response) (cmds []string) {
	for i := 0; i <= len(*responses)-1; i++ {
		cmds = append(cmds, (*responses)[i].Cmd)
	}
	return cmds
}

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

func startRunner(responses *[]tests.Response, port int) (*Runner, error) {
	calls = []string{}
	credentials := StartServer(port, responses)
	var sshExec = exec.NewSSHExecutor(credentials) // Start SSH Connection
	if err := sshExec.Connect(); err != nil {
		return nil, err
	}
	return &Runner{remote: sshExec}, nil
}

func TestChocoExist(t *testing.T) {
	responses := &[]tests.Response{
		{
			Response: "v1.0",
			Error:    nil,
		},
	}
	r, err := startRunner(responses, port+1)
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
	r, err := startRunner(responses, port+1)
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
	r, err := startRunner(responses, port+1)
	assert.Nil(t, err)
	config := v1alpha1.AuxiliarySpec{EnableRDP: &defaultTrue}
	assert.Nil(t, r.EnableRDP(*config.EnableRDP))
}

/*
func TestInstallContainerdSkip(t *testing.T) {
	responses := &[]Response{
		{
			response: "Running",
			error:    nil,
		},
		{
			response: "",
			error:    errors.New("error"),
		},
	}
	r := startRunner(responses)
	config := v1alpha1.ClusterSpec{CalicoVersion: "v3.27"}
	err := r.InstallContainerd(config.CalicoVersion)
	assert.Nil(t, err)
}

func TestInstallContainerdRunning(t *testing.T) {
	responses := &[]Response{
		{
			response: "",
			error:    errors.New(""),
		},
		{
			response: "",
			error:    nil,
		},
	}

	r := startRunner(responses)
	err := r.InstallContainerd("v3.27")
	assert.Nil(t, err)
	assertCalls(t, []string{
		"get-service -name containerd", ".\\Install-Containerd",
	})
}

func TestInstallKubernetes(t *testing.T) {
	responses := &[]Response{
		{
			response: "",
			error:    errors.New(""),
			cmd:      "get-service -name kubelet",
		},
		{
			response: "",
			error:    nil,
			cmd:      ".\\PrepareNode.ps1 -KubernetesVersion v1.29.0",
		},
	}
	cmds := extractCmds(responses)
	r := startRunner(responses)
	err := r.InstallKubernetes("v1.29.0")
	assert.Nil(t, err)
	assertCalls(t, cmds)
}

func TestJoinNodeSkip(t *testing.T) {
	responses := &[]Response{
		{
			response: "Running",
			error:    nil,
			cmd:      "get-service -name kubelet",
		},
	}
	cmds := extractCmds(responses)
	r := startRunner(responses)
	err := r.JoinNode(nil, "v1.29.0", "192.168.0.1")
	assert.Nil(t, err)
	assertCalls(t, cmds)
}

func TestJoinNodeRunner(t *testing.T) {
	responses := &[]Response{
		{
			response: "Stopped",
			error:    nil,
			cmd:      "get-service -name kubelet",
		},
		{
			response: "",
			error:    nil,
			cmd:      "mkdir",
		},
		{
			response: "",
			error:    nil,
			cmd:      "Add-content",
		},
		{
			response: "",
			error:    nil,
			cmd:      "$env:Path",
		},
	}
	cmds := extractCmds(responses)
	r := startRunner(responses)
	fn := func(cmd ...string) (string, error) {
		return "", nil
	}
	err := r.JoinNode(fn, "v1.29.0", "192.168.0.1")
	assert.Nil(t, err)
	assertCalls(t, cmds)
}

func TestInstallCNI(t *testing.T) {
	_ = func(cmd string) (string, error) {
		return "", nil
	}
	//_, _ = SetupRunner{run: fn}
	//err := r.InstallCNI("v3.27.3")
	//assert.Nil(t, err)
}
*/
