package exec

import (
	"bufio"
	"io"
	"os/exec"
	"sync"
)

func Execute(runner interface{}, cmd string, args ...string) (string, error) {
	return runner.(func(cmd string, args ...string) (string, error))(cmd, args...)
}

func RunCommand(cmd string, args ...string) (string, error) {
	command := exec.Command(cmd, args...)

	stdout, err := command.StdoutPipe()
	if err != nil {
		return "", err
	}
	_, err = command.StderrPipe()
	if err != nil {
		return "", err
	}

	// Start and run test command with arguments
	if err := command.Start(); err != nil {
		return "", err
	}

	var output string
	output, err = redirectOutput(nil, stdout)
	if err != nil {
		return "", err
	}

	command.Wait()

	return output, err
}

func redirectOutput(wg *sync.WaitGroup, stdout io.ReadCloser) (string, error) {
	if wg != nil {
		defer wg.Done()
	}
	// Increase max buffer size to 1MB to handle long lines of Ginkgo output and avoid bufio.ErrTooLong errors
	const maxBufferSize = 1024 * 1024
	scanner := bufio.NewScanner(stdout)
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
