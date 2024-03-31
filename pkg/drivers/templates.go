package drivers

import (
	"fmt"
	"log"

	"encoding/xml"
	"libvirt.org/go/libvirt"
)

const domainTmpl = `
<domain type="kvm">
  <name>{{.MachineName}}</name>
  <memory unit="MiB">{{.Memory}}</memory>
  <vcpu placement="static">{{.CPU}}</vcpu>
  <os>
    <type arch="x86_64" machine="pc-q35-7.2">hvm</type>
    <boot dev="hd"/>
  </os>
  <features>
    <acpi/>
    <apic/>
    <hyperv mode="custom">
      <relaxed state="on"/>
      <vapic state="on"/>
      <spinlocks state="on" retries="8191"/>
    </hyperv>
    <vmport state="off"/>
  </features>
  <cpu mode="host-passthrough" check="none" migratable="on"/>
  <clock offset="localtime">
    <timer name="rtc" tickpolicy="catchup"/>
    <timer name="pit" tickpolicy="delay"/>
    <timer name="hpet" present="no"/>
    <timer name="hypervclock" present="yes"/>
  </clock>
  <on_poweroff>destroy</on_poweroff>
  <on_reboot>restart</on_reboot>
  <on_crash>destroy</on_crash>
  <pm>
    <suspend-to-mem enabled="no"/>
    <suspend-to-disk enabled="no"/>
  </pm>
  <devices>
    <emulator>/usr/bin/qemu-system-x86_64</emulator>
    <disk type="file" device="disk">
      <driver name="qemu" type="qcow2"/>
      <source file="{{.DiskPath}}"/>
      <target dev="sda" bus="sata"/>
    </disk>
    <interface type='network'>
      <source network='{{.PrivateNetwork}}'/>
      <model type='virtio'/>
    </interface>
    <interface type="network">
      <source network="{{.Network}}"/>
      <model type="virtio"/>
    </interface>
    <console type="pty">
      <target type="serial" port="0"/>
    </console>
    <input type="tablet" bus="usb">
      <address type="usb" bus="0" port="1"/>
    </input>
    <input type="mouse" bus="ps2"/>
    <input type="keyboard" bus="ps2"/>
    <graphics type="spice" autoport="yes">
      <listen type="address"/>
      <image compression="off"/>
    </graphics>
    <video>
      <model type="qxl" ram="65536" vram="65536" vgamem="16384" heads="1" primary="yes"/>
    </video>
  </devices>
</domain>
`

type kvmIface struct {
	Type string `xml:"type,attr"`
	Mac  struct {
		Address string `xml:"address,attr"`
	} `xml:"mac"`
	Source struct {
		Network string `xml:"network,attr"`
		Portid  string `xml:"portid,attr"`
		Bridge  string `xml:"bridge,attr"`
	} `xml:"source"`
	Target struct {
		Dev string `xml:"dev,attr"`
	} `xml:"target"`
	Model struct {
		Type string `xml:"type,attr"`
	} `xml:"model"`
	Alias struct {
		Name string `xml:"name,attr"`
	} `xml:"alias"`
}

// ifListFromXML returns defined domain interfaces from domain XML.
func ifListFromXML(conn *libvirt.Connect, domain string) ([]kvmIface, error) {
	dom, err := conn.LookupDomainByName(domain)
	if err != nil {
		return nil, fmt.Errorf("failed looking up domain %s: %w", domain, err)
	}
	defer func() { _ = dom.Free() }()

	domXML, err := dom.GetXMLDesc(0)
	if err != nil {
		return nil, fmt.Errorf("failed getting XML of domain %s: %w", domain, err)
	}

	var d struct {
		Interfaces []kvmIface `xml:"devices>interface"`
	}
	err = xml.Unmarshal([]byte(domXML), &d)
	if err != nil {
		return nil, fmt.Errorf("failed parsing XML of domain %s: %w", domain, err)
	}

	return d.Interfaces, nil
}

// macFromXML returns defined MAC address of interface in network from domain XML.
func macFromXML(conn *libvirt.Connect, domain, network string) (string, error) {
	domIfs, err := ifListFromXML(conn, domain)
	if err != nil {
		return "", fmt.Errorf("failed getting network %s interfaces using XML of domain %s: %w", network, domain, err)
	}

	for _, i := range domIfs {
		if i.Source.Network == network {
			log.Println("domain %s has defined MAC address %s in network %s", domain, i.Mac.Address, network)
			return i.Mac.Address, nil
		}
	}

	return "", fmt.Errorf("unable to get defined MAC address of network %s interface using XML of domain %s: network %s not found", network, domain, network)
}
