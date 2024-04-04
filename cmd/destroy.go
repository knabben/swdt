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
	"swdt/pkg/drivers"
	"swdt/pkg/exec"
)

// destroyCmd represents the destroy command
var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Destroy the Windows cluster",
	Long:  "Destroy the Windows cluster.",
	RunE:  RunDestroy,
}

func RunDestroy(cmd *cobra.Command, args []string) error {
	var (
		err    error
		config *v1alpha1.Cluster
	)

	if config, err = loadConfiguration(cmd); err != nil {
		return err
	}

	// Destroy the Windows domain
	if err = destroyWindowsDomain(config); err != nil {
		return err
	}
	// Delete minikube
	if _, err = exec.Execute(exec.RunCommand, "minikube", "delete", "--purge"); err != nil {
		return err
	}
	return nil
}

func destroyWindowsDomain(config *v1alpha1.Cluster) error {
	drv, err := drivers.NewDriver(config)
	if err != nil {
		return err
	}
	return drv.KvmDriver.Remove()
}
