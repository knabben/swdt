package drivers

import (
	"bytes"
	"fmt"
	"log"
	"text/template"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/pkg/errors"
	pkgdrivers "k8s.io/minikube/pkg/drivers"
	"k8s.io/minikube/pkg/drivers/kvm"
	"k8s.io/minikube/pkg/minikube/localpath"
	"libvirt.org/go/libvirt"
)

var (
	network        = "default"
	privateNetwork = "minikube-dev"
)

func NewDriver(diskPath, sshKeyPath string) *kvm.Driver {
	return &kvm.Driver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: "windows",
			StorePath:   localpath.MiniPath(),
			SSHUser:     "Administrator",
			SSHKeyPath:  sshKeyPath,
		},
		Memory:         6000,
		CPU:            4,
		Network:        network,
		PrivateNetwork: privateNetwork,
		DiskPath:       diskPath,
		Hidden:         false,
		NUMANodeCount:  0,
		CommonDriver:   &pkgdrivers.CommonDriver{},
		ConnectionURI:  "qemu:///system",
	}

}

// CreateDomain starts a new libvirt domain from a predefined template.
// copied from Minikube KVM drivers, since we need another template formatted.
func CreateDomain(conn *libvirt.Connect, d *kvm.Driver) (*libvirt.Domain, error) {
	netd, err := conn.LookupNetworkByName(network)
	if err != nil {
		return nil, errors.Wrapf(err, "%s KVM network doesn't exist", network)
	}
	if netd != nil {
		_ = netd.Free()
	}

	// create the XML for the domain using our domainTmpl template
	tmpl := template.Must(template.New("domain").Parse(domainTmpl))
	var domainXML bytes.Buffer
	if err := tmpl.Execute(&domainXML, d); err != nil {
		return nil, errors.Wrap(err, "executing domain xml")
	}

	log.Printf("define libvirt domain using xml: %v\n", domainXML.String())
	// define the domain in libvirt using the generated XML
	dom, err := conn.DomainDefineXML(domainXML.String())
	if err != nil {
		return nil, errors.Wrapf(err, "error defining domain xml: %s", domainXML.String())
	}

	// save MAC address
	dmac, err := macFromXML(conn, d.MachineName, d.Network)
	if err != nil {
		return nil, fmt.Errorf("failed saving MAC address: %w", err)
	}
	d.MAC = dmac
	pmac, err := macFromXML(conn, d.MachineName, d.PrivateNetwork)
	if err != nil {
		return nil, fmt.Errorf("failed saving MAC address: %w", err)
	}
	d.PrivateMAC = pmac

	return dom, nil
}
