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

package connections

import (
	"swdt/apis/config/v1alpha1"
)

type Connection interface {
	// Connect creates the initial connection objects
	Connect() error

	// Run execute the command via transport method
	Run(args string) (string, error)

	// Copy files from and to the node
	Copy(local, remote, perm string) error

	// Close the used connection and sessions
	Close() error
}

func NewConnection(credentials *v1alpha1.SSHSpec) Connection {
	return &SSHConnection{creds: credentials}
}
