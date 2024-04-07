package exec

import (
	"io"
	"os/exec"
	"strings"
	"swdt/pkg/executors/iface"
	"sync"
)

type LocalConnection struct {
	mu     sync.Mutex
	stdout *chan string // Connection stdout channel
	stderr *chan string // Connection stderr channel
}

func (c *LocalConnection) Stdout(std *chan string) {
	c.stdout = std
}

func (c *LocalConnection) Stderr(std *chan string) {
	c.stderr = std
}

func NewLocalExecutor() iface.Executor {
	return &LocalConnection{}
}

func (c *LocalConnection) Run(args string, stdchan *chan string) error {
	var (
		stdout io.Reader
		stderr io.Reader
	)

	// Format local command
	cmd := strings.Split(args, " ")
	command := exec.Command(cmd[0], cmd[1:]...)

	if c.stdout != nil {
		// Send command stdout to stdout channel in not empty.
		stdout, _ = command.StdoutPipe()
		if stdchan == nil {
			stdchan = c.stdout
		}
		go redirectStandard(&c.mu, stdout, stdchan)
	}

	if c.stderr != nil {
		// Send command stdout to stdout channel in not empty.
		stderr, _ = command.StderrPipe()
		go redirectStandard(&c.mu, stderr, c.stderr)
	}

	// Start and run test command with arguments
	return command.Run()
}
