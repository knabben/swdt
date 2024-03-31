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
	"swdt/pkg/drivers"
	"swdt/pkg/exec"

	"github.com/spf13/cobra"
)

// destroyCmd represents the destroy command
var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Destroy the Windows cluster",
	Long:  "Destroy the Windows cluster.",
	RunE:  RunDestroy,
}

func init() {
	destroyCmd.Flags().StringVarP(&kvmQemuURI, "qemu-uri", "q", "qemu:///system", "The KVM QEMU connection URI. (kvm2 driver only)")
}

func RunDestroy(cmd *cobra.Command, args []string) error {
	var err error
	if err = destroyWindowsDomain(); err != nil {
		return err
	}
	if err = exec.Execute(exec.RunCommand, "minikube", "delete", "--purge"); err != nil {
		return err
	}
	return nil
}

func destroyWindowsDomain() error {
	var err error
	if err = drivers.NewDriver(diskPath, sshKey).Remove(); err != nil {
		return err
	}
	return nil

}
