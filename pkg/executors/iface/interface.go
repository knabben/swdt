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

package iface

// Executor is a generic interface executing commands.
type Executor interface {
	// Run execute the command via transport method
	Run(args string, stdout *chan string) error

	// Stdout and stderr iface channel setters
	Stdout(std *chan string)
	Stderr(std *chan string)
}

// SSHExecutor is a interface for SSH connections.
type SSHExecutor interface {
	Executor

	// Copy files from and to the node
	Copy(local, remote, perm string) error

	// Connect creates the initial connection objects
	Connect() error

	// Close the used connection and sessions
	Close() error
}
