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

package exec

import (
	"swdt/pkg/executors/iface"
	"swdt/pkg/executors/tests"
	"testing"

	"github.com/stretchr/testify/assert"
	"swdt/apis/config/v1alpha1"
)

func TestRunWithoutConnect(t *testing.T) {
	credentials := &v1alpha1.SSHSpec{}
	conn := NewSSHExecutor(credentials)
	assert.NotEqual(t, conn, nil)
	err := conn.Run("ls", nil)
	assert.NotNil(t, err)
}

func TestStdoutFromRun(t *testing.T) {
	var cmd = "get-service -name kubelet"
	responses := &[]tests.Response{
		{
			Response: "Running kubelet Kubelet",
			Error:    nil,
		},
	}
	executor := StartServer(t, 2023, responses)
	err := executor.Connect()
	assert.Nil(t, err)

	// make a new stdout channel and set into the runner
	stdout := make(chan string)

	// run the command
	err = executor.Run(cmd, &stdout)
	assert.Nil(t, err)
	var output string
	select {
	case n, ok := <-stdout:
		if ok {
			output = n
		}
	}
	assert.Contains(t, output, "Running")
}

// startServer starts a fake SSH server
func StartServer(t *testing.T, port int, expected *[]tests.Response) iface.SSHExecutor {
	hostname := tests.GetHostname(port)
	tests.NewServer(hostname, expected)
	credentials := &v1alpha1.SSHSpec{
		Hostname: hostname,
		Username: tests.Username,
		Password: tests.FakePassword,
	}
	executor := NewSSHExecutor(credentials)
	assert.NotEqual(t, executor, nil)
	return executor
}
