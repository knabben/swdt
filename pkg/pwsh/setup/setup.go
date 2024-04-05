package setup

import (
	"fmt"
	"github.com/fatih/color"
	klog "k8s.io/klog/v2"
	"strings"
	"swdt/pkg/connections"
	"swdt/pkg/exec"
)

var (
	mainc = color.New(color.FgHiBlack).Add(color.Underline)
	resc  = color.New(color.FgHiGreen).Add(color.Bold)
	bad   = color.New(color.FgHiRed)
)

const (
	CHOCO_PATH    = "C:\\ProgramData\\chocolatey\\bin\\choco.exe"
	CHOCO_INSTALL = "install --accept-licenses --yes"

	cpHost = "control-plane.minikube.internal"
)

type SetupRunner struct {
	conn connections.Connection
	run  func(args string) (string, error)
	copy func(local, remote, perm string) error
}

func (r *SetupRunner) SetConnection(conn *connections.Connection) {
	r.conn = *conn
	r.run = r.conn.Run
	r.copy = r.conn.Copy
}

// InstallChoco proceed to install choco in the default ProgramData folder.
func (r *SetupRunner) InstallChoco() error {
	klog.Info(mainc.Sprint("Installing Choco with PowerShell"))
	if r.ChocoExists() {
		klog.Info(resc.Sprintf("Choco already exists, skipping installation..."))
		return nil
	}
	// Proceed to install choco package manager.
	output, err := r.run(`Set-ExecutionPolicy Bypass -Scope Process -Force;
		[System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; 
		iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))`)
	klog.Info(resc.Sprintf("Installed Choco with: %s", output))
	return err
}

// InstallChocoPackages iterate on a list of packages and execute the installation.
func (r *SetupRunner) InstallChocoPackages(packages []string) error {
	if !r.ChocoExists() {
		return fmt.Errorf("choco not installed. Skipping package installation")
	}
	klog.Info(mainc.Sprint("Installing Choco packages."))
	for _, pkg := range packages {
		output, err := r.run(fmt.Sprintf("%s %s %s", CHOCO_PATH, CHOCO_INSTALL, pkg))
		if err != nil {
			return err
		}
		klog.Info(resc.Sprintf("Installed package %s: %s", pkg, output))
	}
	return nil
}

// ChocoExists check if choco is already installed in the system.
func (r *SetupRunner) ChocoExists() bool {
	_, err := r.run(fmt.Sprintf("%s --version", CHOCO_PATH))
	return err == nil
}

// EnableRDP allow RDP to be accessed in Windows property and Firewall rule
func (r *SetupRunner) EnableRDP(enable bool) error {
	if !enable {
		klog.Warning("Remote Desktop field is disabled. Check the configuration to enable it.")
		return nil
	}
	klog.Info(resc.Sprintf("Enabling RDP."))
	_, err := r.run(`Set-ItemProperty -Path 'HKLM:\System\CurrentControlSet\Control\Terminal Server' -name 'fDenyTSConnections' -value 0; 
		Enable-NetFirewallRule -DisplayGroup 'Remote Desktop'`)
	return err
}

// InstallContainerd install the containerd bits with the set version, enabled required services.
func (r *SetupRunner) InstallContainerd(containerd string) error {
	klog.Info(resc.Sprintf("Installing containerd."))

	// Install containerd if service is not running.
	if output, err := r.run("get-service -name containerd"); err != nil {
		cmd := fmt.Sprintf(".\\Install-Containerd.ps1 -ContainerDVersion %s", containerd)
		output, err := r.run(`curl.exe -LO https://raw.githubusercontent.com/kubernetes-sigs/sig-windows-tools/master/hostprocess/Install-Containerd.ps1; ` + cmd)
		if err != nil {
			return err
		}
		klog.Info(resc.Sprintf("%s -- Containerd v%s installed with success.", output, containerd))
	} else if strings.Contains(output, "Running") { // Otherwise skip
		klog.Info(resc.Sprintf("Skipping containerd installation, service already running, use the copy command."))
	}
	return nil
}

// InstallKubernetes install all Kubernetes bits with the set version.
func (r *SetupRunner) InstallKubernetes(kubernetes string) error {
	klog.Info(resc.Sprintf("Installing kubelet."))
	if output, err := r.run("get-service -name kubelet"); err != nil {
		// Install Kubernetes if service is not running.
		cmd := fmt.Sprintf(".\\PrepareNode.ps1 -KubernetesVersion %s", kubernetes)
		output, err := r.run(`curl.exe -LO https://raw.githubusercontent.com/kubernetes-sigs/sig-windows-tools/master/hostprocess/PrepareNode.ps1; ` + cmd)
		if err != nil {
			return err
		}
		klog.Info(resc.Sprintf("%s -- Kubernetes v%s installed with sucess.", output, kubernetes))
	} else if strings.Contains(output, "kubelet") { // Otherwise skip
		klog.Info(resc.Sprintf("Skipping Kubelet installation, service already running, use the copy command."))
	}
	return nil
}

// JoinNode joins the Windows node into control-plane cluster.
func (r *SetupRunner) JoinNode(cpVersion, cpIPAddr string) error {
	var (
		err    error
		output string
		lout   string
	)

	// In case kubelet is already running, skip joining procedure.
	if output, err = r.run("get-service -name kubelet"); err == nil && !strings.Contains(output, "Running") {
		// Control plane token create and extract, saving the final command
		lcmd := []string{
			"minikube", "ssh", "--", "sudo",
			fmt.Sprintf("/var/lib/minikube/binaries/%s/kubeadm", cpVersion),
			"token", "create", "--print-join-command",
		}
		if lout, err = exec.Execute(exec.RunCommand, lcmd...); err != nil {
			return err
		}
		// Force the creation of the minikube folder for certificates
		cmd := "mkdir c:\\var\\lib\\minikube\\certs -Force"
		if _, err = r.run(cmd); err != nil {
			return err
		}
		// Copy the control plane host value to Windows hosts
		cmd = fmt.Sprintf(`Add-content -Path C:\\Windows\\System32\\drivers\\etc\\hosts -Value \"%s %s\"`, cpIPAddr, cpHost)
		if output, err = r.run(cmd); err != nil {
			return err
		}
		// Trigger a goroutine to copy the ca.crt from kubernetes/pki to the CA folder.
		go func() {
			for {
				klog.Info(resc.Sprintf("trying to copy cert..."))
				cmd := "cp c:\\etc\\kubernetes\\pki\\ca.crt c:\\var\\lib\\minikube\\certs\\ca.crt"
				if _, err = r.run(cmd); err == nil {
					break
				}
			}
		}()
		// Add the control plane into hosts and start the join command.
		cmd = fmt.Sprintf(`$env:Path += ';c:\\k\\'; %s`, strings.Trim(lout, " \n"))
		if output, err = r.run(cmd); err != nil {
			return err
		}
		// Print out node join
		klog.Info(resc.Sprintf(output))

	} else {
		klog.Info(resc.Sprintf("Skipping node join, the Kubelet service is already running."))
	}

	return nil
}

// InstallCNI
func (r *SetupRunner) InstallCNI(calicoVersion string) error {
	// Execute Kubernetes steps for Calico installation
	steps := [][]string{
		{"kubectl", "patch", "ipamconfigurations", "default", "--type", "merge", "--patch='{\"spec\": {\"strictAffinity\": true}}'"},

		/*{"kubectl", "config", "set-context", "minikube"},
		{"kubectl", "create", "-f", fmt.Sprintf("https://raw.githubusercontent.com/projectcalico/calico/%v/manifests/tigera-operator.yaml", calicoVersion)},

		// Render the configmap with content
		{"kubectl", "create", "-f", "./specs/configmap.yaml"},
		{"kubectl", "create", "-f", "./specs/installation.yaml"},
		{"kubectl", "create", "-f", "./specs/apiserver.yaml"},
		*/
	}

	//{curl -L  https://raw.githubusercontent.com/kubernetes-sigs/sig-windows-tools/master/hostprocess/calico/kube-proxy/kube-proxy.yml | sed "s/KUBE_PROXY_VERSION/v1.27.3/g" | kubectl apply -f -
	for i := 0; i <= len(steps)-1; i++ {
		resc.Printf("Running: %v\n", strings.Join(steps[i], " "))
		if stdout, err := exec.Execute(exec.RunCommand, steps[i]...); err != nil {
			bad.Println(err.Error())
		} else {
			fmt.Println(stdout)
		}
	}
	return nil
}
