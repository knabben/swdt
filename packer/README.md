## Packer VM image builder

This folder hosts the plain boot and automatic installation scripts
using packer, the final outcome is the qemu artifact ready to be used
as a VM for swdt with SSH enabled.

Pre-requisites:

* Hashicorp Packer >=1.10.2

2 ISOs are required, save them on isos folder:

* **window.iso** - [Windows 2022 Server](https://www.microsoft.com/en-us/evalcenter/evaluate-windows-server-2022) 
* **virtio.iso** - [Windows Virtio Drivers](https://fedorapeople.org/groups/virt/virtio-win/direct-downloads/archive-virtio/virtio-win-0.1.240-1/virtio-win-0.1.240.iso)

A SSH key is required to exist on `~/.ssh/id_rsa.pub`, changeable in the Makefile.

### Running 

```shell
make start
```

Behind the scenes it will call Packer in the kvm build

```shell
packer init kvm
PACKER_LOG=1 packer build kvm
```

### Export

The folder `output` will contain the `win2k22` QEMU QCOW Image.

