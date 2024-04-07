packer {
  required_plugins {
    qemu = {
      version = "~> 1"
      source  = "github.com/hashicorp/qemu"
    }
  }
}

variable "windows_iso" {
  type = string
}

variable "virtio_iso" {
  type = string
}

variable "windows_sha256" {
  type = string
}

source "qemu" "windows" {
  vm_name     = "windows"
  format      = "qcow2"
  accelerator = "kvm"

  iso_url      = var.windows_iso
  iso_checksum = var.windows_sha256

  cpus   = 4
  memory = 4096

  efi_boot       = false
  disk_size      = "15G"
  disk_interface = "virtio"

  floppy_files = [
    "packer/kvm/floppy/autounattend.xml",
    "packer/kvm/floppy/bootstrap.ps1",
    "packer/kvm/floppy/ssh_key.pub"
  ]
  qemuargs = [["-cdrom", var.virtio_iso]]

  output_directory = "packer/output"

  communicator = "ssh"
  ssh_username = "Administrator"
  ssh_password = "S3cr3t0!"
  ssh_timeout  = "1h"

  boot_wait        = "10s"
  shutdown_command = "shutdown /s /t 120 /f"
  shutdown_timeout = "15m"
}

build {
  name    = "win2022"
  sources = ["source.qemu.windows"]
}

