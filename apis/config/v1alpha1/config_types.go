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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SSHSpec struct {
	// Username set the Windows user
	Username string `json:"username,omitempty"`

	// Hostname set the Windows node endpoint
	Hostname string `json:"hostname,omitempty"`

	// Password is the SSH password for this user
	Password string `json:"password,omitempty"`

	// PrivateKey is the SSH private path for this user
	PrivateKey string `json:"privateKey,omitempty"`
}

type VirtualizationSpec struct {
	// KVM Qemu URI is the path of qemu socket URI
	KvmQemuURI string `json:"kvmQemuURI,omitempty"`
	// DiskPath is the path of the Windows qcow2 file.

	DiskPath string `json:"diskPath,omitempty"`

	// SSH stored the Windows VM credentials.
	SSH *SSHSpec `json:"ssh,omitempty"`
}

type AuxiliarySpec struct {
	// EnableRDP set up the remote desktop service and enable firewall for it.
	EnableRDP *bool `json:"enableRDP"`
	// ChocoPackages provides a list of packages automatically installed in the node.
	ChocoPackages *[]string `json:"chocoPackages,omitempty"`
}

type ProvisionerSpec struct {
	// Name of the service to be deployed
	Name string `json:"name,omitempty"`

	// Version is the binary version to be deployed
	Version string `json:"version,omitempty"`

	// SourceURL set the HTTP server to be downloaded from
	SourceURL string `json:"sourceURL,omitempty"`

	// Destination set the Windows patch to upload the file
	Destination string `json:"destination,omitempty"`

	// Overwrite delete the old file if exists first.
	Overwrite bool `json:"overwrite,omitempty"`
}

// WorkloadSpec defines the workload specification
type WorkloadSpec struct {
	// Virtualization defines libvirt configuration.
	Virtualization VirtualizationSpec `json:"virtualization,omitempty"`

	// Auxiliary defines the specification for 3rd party procedures in the node.
	Auxiliary AuxiliarySpec `json:"auxiliary,omitempty"`

	// Provisioners defines the binaries installations
	Provisioners []ProvisionerSpec `json:"provisioners"`
}

// ControlPlaneSpec defines the control plane specification
type ControlPlaneSpec struct {
	// Minikube set the control plane installation via minikube
	// otherwise the kubeconfig is being used
	Minikube bool `json:"minikube,omitempty"`
}

// ClusterSpec defines the desired state of the Cluster
type ClusterSpec struct {
	// Version is the Kubernetes version being installed in the cluster
	Version string `json:"version,omitempty"`

	ControlPlane *ControlPlaneSpec `json:"controlPlane,omitempty"`
	Workload     *WorkloadSpec     `json:"workload,omitempty"`
}

// ClusterStatus -- tbd
type ClusterStatus struct{}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+k8s:defaulter-gen=true

// Cluster is the Schema for the configuration API
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec,omitempty"`
	Status ClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ClusterList contains a list of Node
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cluster{}, &ClusterList{})
}
