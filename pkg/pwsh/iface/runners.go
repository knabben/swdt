package iface

import (
	"swdt/apis/config/v1alpha1"
	"swdt/pkg/executors/exec"
	"swdt/pkg/executors/iface"
	"swdt/pkg/pwsh/kubernetes"
	"swdt/pkg/pwsh/setup"
)

type Runner[R RunnerInterface] struct {
	Inner R
}

type RunnerInterface interface {
	*setup.Runner | *kubernetes.Runner
	SetLocal(executor iface.Executor)
	SetRemote(executor iface.SSHExecutor)
}

// NewRunner returns the encapsulated picked runner and sets its executors
func NewRunner[R RunnerInterface](ssh *v1alpha1.SSHSpec, run R) (*Runner[R], error) {
	var sshExec = exec.NewSSHExecutor(ssh) // Start SSH Connection
	run.SetRemote(sshExec)
	run.SetLocal(exec.NewLocalExecutor()) // Start Local executor
	if err := sshExec.Connect(); err != nil {
		return nil, err
	}
	return &Runner[R]{Inner: run}, nil
}
