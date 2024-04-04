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
	"log"
	"strings"
	"swdt/apis/config/v1alpha1"

	"swdt/pkg/drivers"
	"swdt/pkg/exec"

	"github.com/spf13/cobra"
	"libvirt.org/go/libvirt"
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
		err    error
		config *v1alpha1.Cluster
	)

	if config, err = loadConfiguration(cmd); err != nil {
		return err
	}

	// Start the minikube if the flag is enabled.
	if config.Spec.ControlPlane.Minikube {
		version := config.Spec.ControlPlane.KubernetesVersion
		if err := startMinikube(exec.RunCommand, version); err != nil {
			return err
		}
	}

	// Start the Windows VM on LibVirt
	return startWindowsVM(config)
}

// startWindowsVM create the Windows libvirt domain and start it.
func startWindowsVM(config *v1alpha1.Cluster) error {
	var (
		dom *libvirt.Domain
		err error
	)

	drv, err := drivers.NewDriver(config)
	if err != nil {
		return err
	}

	log.Println("Creating domain...")

	// Create the libvirt domain
	if dom, err = drv.CreateDomain(); err != nil {
		// Domain already exists, skipping the Windows creation.
		if alreadyExists(err) {
			return nil
		}
		return err
	}
	defer func() {
		if ferr := dom.Free(); ferr != nil {
			log.Printf("unable to free domain: %v\n", err)
			err = ferr
		}
	}()

	// Start the Windows created domain.
	if err = drv.KvmDriver.Start(); err != nil {
		return err
	}

	return nil
}

// startMinikube initialize a minikube control plane.
func startMinikube(executor interface{}, version string) (err error) {
	// Start minikube with KVM2 machine
	args := []string{
		"start", "--driver", "kvm2",
		"--container-runtime", "containerd",
		"--kubernetes-version", version,
	}
	if err = exec.Execute(executor, "minikube", args...); err != nil {
		return err
	}
	return nil
}

func alreadyExists(err error) bool {
	return strings.Contains(err.Error(), "already exists with")
}
