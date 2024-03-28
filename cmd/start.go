/*
Copyright 2024 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"os/exec"
	"sync"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starting the Windows cluster end to end. (idoubt)",
	Long:  "Starting the Windows cluster end to end.",
	RunE:  RunStart,
}

func RunStart(cmd *cobra.Command, args []string) error {
	var (
		err error
	)

	// Delete minikube if already exists.
	if err = execute("minikube", "delete", "--purge"); err != nil {
		return err
	}

	// Start minikube with KVM2 machine
	if err = execute("minikube", "start", "--driver", "kvm2"); err != nil {
		return err
	}

	return nil

}

func execute(cmd string, args ...string) (err error) {
	command := buildCmd(cmd, args...)

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

	go redirectOutput(&wg, stdout)
	redirectOutput(nil, stderr)

	wg.Wait()
	return command.Wait()

	return err
}

func buildCmd(cmd string, args ...string) *exec.Cmd {
	return exec.Command(cmd, args...)
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
