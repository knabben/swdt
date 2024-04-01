# Design document for the SWDT CLI

<img src=".assets/logo.png" width="250" height="250">

This document is a reference for the *Wave (Windows Automated Virtual Environment)* decisions regarding multi-OS support architecture. It is developed with Cobra/spf13 to conform to a stable namespace and different connection libraries to support remote command execution.

Follow the [docs](../../docs) for more information.

## Subcommands and namespace

These are the current supported subcommands for the program:

* `swdt start`
  * Initialize the entire Cluster, the control plane is supported using Minikube, the Windows node is started by Libvirt.  
* `swdt aux`
  * Initialize the node auxiliary tools and procedures like enabling RDP, installing Choco and packages, etc.
* `swdt copy`
  * Deploy Kubernetes binaries from the HTTP server indicated in the configuration.
* `swdt readiness`
  * Run the [windows operational readiness](https://github.com/kubernetes-sigs/windows-operational-readiness) project in the local cluster

## Configuration

The configuration API follows the GVK (GroupVersionKind) model from Kubernetes, using api-machinery for marshaling and unmarshalling, as well as defaulting values and validating its content. The goal of reusing this API is to enable sharing of the data structure not only with the CLI, but also with controllers and other projects in a well-known and agreed-upon format.

The project contains a Makefile with targets to generate the required API management and boilerplate functions

```mermaid
flowchart TD

    U(Admin) -->|writes CR| A(Kind: Config)
    A(Kind: Config) -->|parsed by| B(CLI / decoder)
    A(Kind: Config) -->|parsed by| C(Controller)

    B(CLI / decoder) -->|used by| D(business logic / pkg)
    C(Controller) -->|used by| D(business logic / pkg)
```

### Fields and schema

These are the supported fields for configuring access node credentials, setting up options, and providing Kubernetes binaries information. 

* Credentials: support for SSH and WinRM settings
* Setup: Define options for initial node bootstrap
* Kubernetes: Auxiliary Kubernetes binaries installation

Following a configuration sample:

```
apiVersion: windows.k8s.io/v1alpha1
kind: Node
metadata:
  name: sample
spec:
  credentials:
    username: Administrator
    hostname: 192.168.122.220:22
    publicKey:
  setup:
    enableRDP: true
    chocoPackages:
      - vim
      - grep
  kubernetes:
    provisioners:
      - name: containerd:
        version: 1.7.11
        sourceURL: http://xyz/containerd.exe
        destination: c:\Program Files\containerd\bin\containerd.exe
        overwrite: true
      - name: kubelet
        version: 1.29.0
        sourceURL: http://xyz/kubelet.exe
        destination: c:\k\kubelet.exe
        overwrite: true
```

## Connections

Currently, the project SSH for running commands remotely on the node. The common fields required are username and hostname. To proceed, ssh object content should be filled out with the proper connections parameters.

## Testing

See [experimental early guide for testers](samples/mloskot/README.windows.md)
dedicated to try SWDT CLI on Windows host.
