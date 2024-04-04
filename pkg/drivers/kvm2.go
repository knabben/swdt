package drivers

import (
	"bytes"
	"fmt"
	"log"
	"swdt/apis/config/v1alpha1"
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
	privateNetwork = "mk-minikube"

	windowsDomain = "windows"
	hostName      = "win2k22"
)

type Driver struct {
	KvmDriver *kvm.Driver
	Conn      *libvirt.Connect
}

func NewDriver(config *v1alpha1.Cluster) (*Driver, error) {
	uri := config.Spec.Workload.Virtualization.KvmQemuURI
	conn, err := libvirt.NewConnect(uri)
	if err != nil {
		return nil, err
	}
	return &Driver{
		KvmDriver: &kvm.Driver{
			BaseDriver: &drivers.BaseDriver{
				MachineName: windowsDomain,
				StorePath:   localpath.MiniPath(),
				SSHUser:     "Administrator",
				SSHKeyPath:  config.Spec.Workload.Virtualization.SSH.PrivateKey,
			},
			Memory:         6000,
			CPU:            4,
			Network:        network,
			PrivateNetwork: privateNetwork,
			DiskPath:       config.Spec.Workload.Virtualization.DiskPath,
			Hidden:         false,
			NUMANodeCount:  0,
			CommonDriver:   &pkgdrivers.CommonDriver{},
			ConnectionURI:  uri,
		},
		Conn: conn,
	}, nil
}

// CreateDomain starts a new libvirt domain from a predefined template.
// copied from Minikube KVM drivers, since we need another template formatted.
func (d *Driver) CreateDomain() (*libvirt.Domain, error) {
	netd, err := d.Conn.LookupNetworkByName(network)
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
	dom, err := d.Conn.DomainDefineXML(domainXML.String())
	if err != nil {
		return nil, errors.Wrapf(err, "error defining domain xml: %s", domainXML.String())
	}

	// save MAC address
	dmac, err := macFromXML(d.Conn, d.KvmDriver.MachineName, d.KvmDriver.Network)
	if err != nil {
		return nil, fmt.Errorf("failed saving MAC address: %w", err)
	}
	d.KvmDriver.MAC = dmac
	pmac, err := macFromXML(d.Conn, d.KvmDriver.MachineName, d.KvmDriver.PrivateNetwork)
	if err != nil {
		return nil, fmt.Errorf("failed saving MAC address: %w", err)
	}
	d.KvmDriver.PrivateMAC = pmac

	return dom, nil
}

// GetPrivWindowsIPAddress returns the PRIVATE ip address for Windows domain
func (d *Driver) GetPrivWindowsIPAddress() (string, error) {
	var (
		networks []libvirt.Network
		err      error
	)
	networks, err = d.Conn.ListAllNetworks(libvirt.CONNECT_LIST_NETWORKS_ACTIVE)
	if err != nil {
		return "", err
	}
	for _, network := range networks {
		name, err := network.GetName()
		if err != nil {
			return "", err
		}
		if name == privateNetwork {
			leases, err := network.GetDHCPLeases()
			if err != nil {
				return "", err
			}
			for _, lease := range leases {
				if lease.Hostname == hostName {
					return lease.IPaddr, nil
				}
			}
		}
	}
	return "", errors.New("No lease was found.")
}
