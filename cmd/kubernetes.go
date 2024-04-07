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
	"github.com/spf13/cobra"
	"swdt/apis/config/v1alpha1"
	ifacer "swdt/pkg/pwsh/iface"
	"swdt/pkg/pwsh/kubernetes"
)

// setupCmd represents the setup command
var kubernetesCmd = &cobra.Command{
	Use:   "kubernetes",
	Short: "Provision Kubernetes binaries into a running node",
	Long:  `Provision Kubernetes binaries into a running node`,
	RunE:  RunKubernetes,
}

func RunKubernetes(cmd *cobra.Command, args []string) error {
	var (
		err    error
		config *v1alpha1.Cluster
	)
	if config, err = loadConfiguration(cmd); err != nil {
		return err
	}

	// Starting the executor
	ssh := config.Spec.Workload.Virtualization.SSH
	r, err := ifacer.NewRunner(ssh, &kubernetes.Runner{})
	if err != nil {
		return err
	}

	return r.Inner.InstallProvisioners(config.Spec.Workload.Provisioners)
}
