package exec

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"sync"
)

func Execute(runner interface{}, cmd string, args ...string) error {
	return runner.(func(cmd string, args ...string) error)(cmd, args...)
}

func RunCommand(cmd string, args ...string) error {
	command := exec.Command(cmd, args...)

	stdout, err := command.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := command.StderrPipe()
	if err != nil {
		return err
	}

	// Start and run test command with arguments
	if err := command.Start(); err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		err := redirectOutput(&wg, stdout)
		if err != nil {
		}
	}()
	if err = redirectOutput(nil, stderr); err != nil {
		return err
	}

	wg.Wait()
	command.Wait()
	return nil
}

func redirectOutput(wg *sync.WaitGroup, stdout io.ReadCloser) error {
	if wg != nil {
		defer wg.Done()
	}
	// Increase max buffer size to 1MB to handle long lines of Ginkgo output and avoid bufio.ErrTooLong errors
	const maxBufferSize = 1024 * 1024
	scanner := bufio.NewScanner(stdout)
	buf := make([]byte, 0, maxBufferSize)
	scanner.Buffer(buf, maxBufferSize)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}
