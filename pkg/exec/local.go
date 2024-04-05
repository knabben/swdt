package exec

import (
	"bufio"
	"github.com/pkg/errors"
	"io"
	"os/exec"
	"sync"
)

func Execute(runner interface{}, cmd ...string) (string, error) {
	return runner.(func(cmd ...string) (string, error))(cmd...)
}

func RunCommand(cmd ...string) (string, error) {
	var (
		stdoutp string
		stderrp string
	)
	command := exec.Command(cmd[0], cmd[1:]...)

	stdout, err := command.StdoutPipe()
	if err != nil {
		return "", err
	}
	stderr, err := command.StderrPipe()
	if err != nil {
		return "", err
	}

	// Start and run test command with arguments
	if err := command.Start(); err != nil {
		return "", err
	}
	stdoutp, err = redirectOutput(nil, stdout)
	if err != nil {
		return "", err
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func(closer *io.ReadCloser, wg *sync.WaitGroup) {
		if stderrp, err = redirectOutput(wg, *closer); err == nil && len(stderrp) > 0 {
			err = errors.New(stderrp)
		}
	}(&stderr, &wg)
	wg.Wait()
	command.Wait()
	return stdoutp, err
}

func redirectOutput(wg *sync.WaitGroup, std io.ReadCloser) (string, error) {
	if wg != nil {
		defer wg.Done()
	}
	// Increase max buffer size to 1MB to handle long lines of Ginkgo output and avoid bufio.ErrTooLong errors
	const maxBufferSize = 1024 * 1024
	scanner := bufio.NewScanner(std)
	buf := make([]byte, 0, maxBufferSize)
	scanner.Buffer(buf, maxBufferSize)
	scanner.Split(bufio.ScanLines)
	var output string
	for scanner.Scan() {
		output += scanner.Text() + "\n"
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return output, nil
}
