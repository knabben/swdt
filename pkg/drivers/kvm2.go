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
	DefaultNetwork = "default"
	PrivateNetwork = "mk-minikube"

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
			Network:        DefaultNetwork,
			PrivateNetwork: PrivateNetwork,
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
	netd, err := d.Conn.LookupNetworkByName(DefaultNetwork)
	if err != nil {
		return nil, errors.Wrapf(err, "%s KVM network doesn't exist", DefaultNetwork)
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

// GetLeasedIPs returns the network IP leases address for all domains
func (d *Driver) GetLeasedIPs(filterNetwork string) (leases map[string]string, err error) {
	var networks []libvirt.Network
	networks, err = d.Conn.ListAllNetworks(libvirt.CONNECT_LIST_NETWORKS_ACTIVE)
	if err != nil {
		return leases, err
	}
	leases = make(map[string]string, len(networks))
	for _, network := range networks {
		name, err := network.GetName()
		if err != nil {
			return leases, err
		}
		if name == filterNetwork {
			dhcpLeases, err := network.GetDHCPLeases()
			if err != nil {
				return leases, err
			}
			for _, lease := range dhcpLeases {
				leases[lease.Hostname] = lease.IPaddr
			}
		}
	}
	return leases, nil
}
