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
		drv, err := drivers.NewDriver(config)
		if err != nil {
			return err
		}

		ipAddr, err := drv.GetPrivWindowsIPAddress()
		if err != nil {
			return err
		}

		config.Spec.Workload.Virtualization.SSH.Hostname = ipAddr
	}

	runner, err := executor.NewRunner(config, &setup.SetupRunner{})
	if err != nil {
		return err
	}
	defer runner.CloseConnection() // nolint

	// Install choco binary
	if err = runner.Inner.InstallChoco(); err != nil {
		return err
	}

	/*	// Install Choco packages from the input list
		if err = runner.Inner.InstallChocoPackages(*config.Spec.Workload.Auxiliary.ChocoPackages); err != nil {
			return err
		}

		// Enable RDP if the option is true
		return runner.Inner.EnableRDP(*config.Spec.Workload.Auxiliary.EnableRDP)d
	*/
	return nil
}
