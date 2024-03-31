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
	"github.com/pkg/errors"
	"log"
	"path/filepath"
	"strings"

	"swdt/pkg/drivers"
	"swdt/pkg/exec"

	"github.com/docker/docker/pkg/homedir"
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

var (
	diskPath        string
	sshKey          string
	triggerMinikube bool
	kvmQemuURI      string
)

func init() {
	startCmd.Flags().StringVarP(&kvmQemuURI, "qemu-uri", "q", "qemu:///system", "The KVM QEMU connection URI. (kvm2 driver only)")
	startCmd.Flags().StringVarP(&sshKey, "ssh-key", "s", filepath.Join(homedir.Get(), ".ssh/id_rsa"), "The KVM QEMU connection URI. (kvm2 driver only)")
	startCmd.Flags().StringVarP(&diskPath, "disk-path", "d", "", "Windows qcow2 disk path.")
	startCmd.Flags().BoolVarP(&triggerMinikube, "minikube", "m", false, "Trigger minikube installation")
}

func RunStart(cmd *cobra.Command, args []string) error {
	var err error
	if err = validateFlags(); err != nil {
		return err
	}
	if triggerMinikube {
		if err := startMinikube(exec.RunCommand); err != nil {
			return err
		}
	}
	// Start the Windows VM on LibVirt
	return startWindowsVM()
}

func validateFlags() error {
	if diskPath == "" {
		return errors.New("No disk path is passed. Use the disk-path argument to bring the Windows image.")
	}
	return nil
}

// startWindowsVM create the Windows libvirt domain and start it.
func startWindowsVM() error {
	var (
		conn *libvirt.Connect
		dom  *libvirt.Domain
		err  error
	)

	// Start the Libvirt connection with kvm Qemu URI
	if conn, err = libvirt.NewConnect(kvmQemuURI); err != nil {
		return err
	}
	log.Println("Creating domain...")
	drv := drivers.NewDriver(diskPath, sshKey)

	// Create the libvirt domain
	if dom, err = drivers.CreateDomain(conn, drv); err != nil {
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
	if err = drv.Start(); err != nil {
		return err
	}
	return nil
}

func alreadyExists(err error) bool {
	return strings.Contains(err.Error(), "already exists with")
}

// startMinikube initialize a minikube control plane.
func startMinikube(executor interface{}) (err error) {
	// Start minikube with KVM2 machine
	if err = exec.Execute(executor, "minikube", "start", "--driver", "kvm2"); err != nil {
		return err
	}
	return nil
}
