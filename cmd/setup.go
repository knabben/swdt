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
	"swdt/apis/config/v1alpha1"
	"swdt/pkg/drivers"
	"swdt/pkg/pwsh/executor"
	"swdt/pkg/pwsh/setup"

	"github.com/spf13/cobra"
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

	// Find the IP of the Windows machine grabbing from the domain
	if config.Spec.Workload.Virtualization.SSH.Hostname == "" {
		var ipAddr string
		if ipAddr, err = findWindowsIP(config); err != nil {
			return err
		}
		config.Spec.Workload.Virtualization.SSH.Hostname = ipAddr
	}

	runner, err := executor.NewRunner(config, &setup.SetupRunner{})
	if err != nil {
		return err
	}
	defer func(runner *executor.Runner[*setup.SetupRunner]) {
		if err := runner.CloseConnection(); err != nil {
			log.Printf("error to close the connection: %v\n", err)
		}
	}(runner)

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
	// Enable RDP if option is true
	if err = runner.Inner.EnableRDP(*config.Spec.Workload.Auxiliary.EnableRDP); err != nil {
		return err
	}

	return runner.Inner.InstallContainerd(config.Spec.Workload.ContainerdVersion)
}

func findWindowsIP(config *v1alpha1.Cluster) (string, error) {
	drv, err := drivers.NewDriver(config)
	if err != nil {
		return "", err
	}
	var ipAddr string
	if ipAddr, err = drv.GetPrivWindowsIPAddress(); err != nil {
		return "", err
	}
	return ipAddr, err
}
