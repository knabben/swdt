WINDOWS_ISO = './isos/windows.iso'
WINDOWS_SHA256 = 'sha256:3e4fa6d8507b554856fc9ca6079cc402df11a8b79344871669f0251535255325'
VIRTIO_ISO = '${PWD}/isos/virtio.iso'
KEY_PATH = '${HOME}/.ssh/id_rsa.pub'

start:
	rm -fr output
	packer init kvm
	ln -s $(KEY_PATH) ./kvm/floppy/ssh_key.pub || true
	PACKER_LOG=1 packer build \
		-var 'windows_iso=$(WINDOWS_ISO)' \
		-var 'windows_sha256=$(WINDOWS_SHA256)' \
		-var 'virtio_iso=$(VIRTIO_ISO)' \
		kvm/
