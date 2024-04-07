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
	"github.com/fatih/color"
	"k8s.io/klog/v2"
	"log"
	"strings"
	"swdt/apis/config/v1alpha1"

	"github.com/spf13/cobra"
	"libvirt.org/go/libvirt"
	"swdt/pkg/drivers"
)

var (
	resc = color.New(color.FgHiGreen).Add(color.Bold)
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

	// Loading configuration
	if config, err = loadConfiguration(cmd); err != nil {
		return err
	}

	// Start the minikube if the flag is enabled.
	if config.Spec.ControlPlane.Minikube {
		//version := config.Spec.ControlPlane.KubernetesVersion
		klog.Info(resc.Sprintf("Starting a Minikube control plane, this operation can take a while..."))
		/*if err := startMinikube(exec.RunCommand, version); err != nil {
			return err
		}*/
	}

	// Start the Windows VM on LibVirt
	klog.Info(resc.Sprintf("Starting the Windows VM on domain, this operation can take a while..."))
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
	/*
		cmd := []string{
			"minikube", "start", "--driver", "kvm2",
			"--container-runtime", "containerd",
			"--kubernetes-version", version,
		}
		if _, err = exec.Execute(executor, cmd...); err != nil {
			return err
		}*/
	return nil
}

func alreadyExists(err error) bool {
	return strings.Contains(err.Error(), "already exists with")
}
