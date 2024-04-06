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
	"k8s.io/klog/v2"
	"log"

	"github.com/spf13/cobra"
	"swdt/apis/config/v1alpha1"
	"swdt/pkg/drivers"
	"swdt/pkg/pwsh/executor"
	"swdt/pkg/pwsh/setup"
)

var (
	windowsHost      = "windows"
	controlPlaneHost = "minikube"
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Bootstrap the node via basic unit setup",
	Long:  `Bootstrap the node via basic unit setup`,
	RunE:  RunSetup,
}

func RunSetup(cmd *cobra.Command, args []string) error {
	var (
		err    error
		config *v1alpha1.Cluster
	)

	if config, err = loadConfiguration(cmd); err != nil {
		return err
	}

	var leases map[string]string
	if leases, err = findPrivateIPs(config); err != nil {
		return err
	}
	// Find the IP of the Windows machine grabbing from the domain
	if config.Spec.Workload.Virtualization.SSH.Hostname == "" {
		config.Spec.Workload.Virtualization.SSH.Hostname = leases[windowsHost] + ":22"
	}
	// Find the control plane IP
	_ = leases[controlPlaneHost]
	klog.Info(resc.Sprintf("Found DHCP leases: %v", leases))

	// Starting the SSH executor
	runner, err := executor.NewRunner(config, &setup.SetupRunner{})
	if err != nil {
		return err
	}
	defer func(runner *executor.Runner[*setup.SetupRunner]) {
		if err := runner.CloseConnection(); err != nil {
			log.Printf("error to close the connection: %v\n", err)
		}
	}(runner)

	/*
		// Install choco binary and packages if a list of packages exists
		if len(*config.Spec.Workload.Auxiliary.ChocoPackages) > 0 {
			if err = runner.Inner.InstallChoco(); err != nil {
				return err
			}
			// Install Choco packages from the input list
			if err = runner.Inner.InstallChocoPackages(*config.Spec.Workload.Auxiliary.ChocoPackages); err != nil {
				return err
			}
		}
	*/

	// Enable RDP if option is true
	rdp := config.Spec.Workload.Auxiliary.EnableRDP
	if err = runner.Inner.EnableRDP(*rdp); err != nil {
		return err
	}

	/*
		// Installing Containerd with predefined version
		containerd := config.Spec.Workload.ContainerdVersion
		if err = runner.Inner.InstallContainerd(containerd); err != nil {
			return err
		}

		// Installing Kubeadm and Kubelet binaries in the host
		kubernetes := config.Spec.Workload.KubernetesVersion
		if err = runner.Inner.InstallKubernetes(kubernetes); err != nil {
			return err
		}

		// Joining the Windows node in the control plane.
		cpKubernetes := config.Spec.ControlPlane.KubernetesVersion
		if err = runner.Inner.JoinNode(exec.RunCommand, cpKubernetes, controlPlaneIP); err != nil {
			return err
		}
	*/
	// Install Calico CNI operator and CR
	// NOTE: Only Calico is supported for now on HPC
	if err = runner.Inner.InstallCNI(config.Spec.CalicoVersion); err != nil {
		return err
	}

	return nil
}

// findPrivateIPs returns the leased ips from the domain.
func findPrivateIPs(config *v1alpha1.Cluster) (leases map[string]string, err error) {
	var drv *drivers.Driver
	drv, err = drivers.NewDriver(config)
	if err != nil {
		return
	}
	if leases, err = drv.GetLeasedIPs(drivers.PrivateNetwork); err != nil {
		return
	}
	return
}
