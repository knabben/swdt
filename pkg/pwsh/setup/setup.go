package setup

import (
	"fmt"
	"github.com/fatih/color"
	"k8s.io/klog/v2"
	"strings"
	"swdt/pkg/executors/exec"
	"swdt/pkg/executors/iface"
	"swdt/pkg/templates"
	"time"
)

var (
	mainc = color.New(color.FgHiYellow).Add(color.Underline)
	resc  = color.New(color.FgHiGreen).Add(color.Bold)
	warn  = color.New(color.FgWhite)
	bad   = color.New(color.FgHiRed)
)

const (
	CHOCO_PATH    = "C:\\ProgramData\\chocolatey\\bin\\choco.exe"
	CHOCO_INSTALL = "install --accept-licenses --yes"

	cpHost = "control-plane.minikube.internal"
)

type Runner struct {
	Logging bool // enabled verbose logging on calls (both stdout and stderr)
	remote  iface.SSHExecutor
	local   iface.Executor
}

func (r *Runner) SetLocal(executor iface.Executor) {
	r.local = executor
}

func (r *Runner) SetRemote(executor iface.SSHExecutor) {
	r.remote = executor
}

// runL runs a local command using the local executor
func (r *Runner) runL(args string) error {
	return r.local.Run(args, nil)
}

// runLstd runs a local command using the local executor
func (r *Runner) runLstd(args string, stdout *chan string) error {
	return r.local.Run(args, stdout)
}

// runR runs a local command using the local executor
func (r *Runner) runR(args string) error {
	return r.remote.Run(args, nil)
}

// runRstdout runs a local command using the local executor
func (r *Runner) runRstd(args string, stdout *chan string) error {
	return r.remote.Run(args, stdout)
}

// ChocoExists check if choco is already installed in the system.
func (r *Runner) ChocoExists() bool {
	return r.runR(fmt.Sprintf("%s --version", CHOCO_PATH)) == nil
}

// InstallChoco proceed to install choco in the default ProgramData folder.
func (r *Runner) InstallChoco() error {
	klog.Info(mainc.Sprint("Installing Choco with PowerShell."))

	if r.Logging {
		go exec.EnableOutput(nil, r.remote.Stdout)
		go exec.EnableOutput(nil, r.remote.Stderr)
	}

	if r.ChocoExists() {
		klog.Info(resc.Sprintf("Choco already exists, skipping installation..."))
		return nil
	}
	return r.runR(`Set-ExecutionPolicy Bypass -Scope Process -Force;
		[System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072;
		iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))`)
}

// InstallChocoPackages iterate on a list of packages and execute the installation.
func (r *Runner) InstallChocoPackages(packages []string) error {
	if !r.ChocoExists() {
		return fmt.Errorf("choco not installed. Skipping package installation")
	}

	klog.Info(mainc.Sprint("Installing Choco packages."))
	for _, pkg := range packages {
		if err := r.runR(fmt.Sprintf("%s %s %s", CHOCO_PATH, CHOCO_INSTALL, pkg)); err != nil {
			return err
		}
	}
	return nil
}

// EnableRDP allow RDP to be accessed in Windows property and Firewall rule
func (r *Runner) EnableRDP(enable bool) error {
	if !enable {
		klog.Warning("Remote Desktop field is disabled. Check the configuration to enable it.")
		return nil
	}

	if r.Logging {
		go exec.EnableOutput(nil, r.remote.Stdout)
		go exec.EnableOutput(nil, r.remote.Stderr)
	}

	klog.Info(mainc.Sprint("Enabling Remote Desktop."))
	return r.runR(`Set-ItemProperty -Path 'HKLM:\System\CurrentControlSet\Control\Terminal Server' -name 'fDenyTSConnections' -value 0;
		Enable-NetFirewallRule -DisplayGroup 'Remote Desktop'`)
}

// InstallContainerd install the containerd bits with the set version, enabled required services.
func (r *Runner) InstallContainerd(containerd string) error {
	klog.Info(mainc.Sprintf("Installing containerd."))

	var output string
	if r.Logging {
		go exec.EnableOutput(nil, r.remote.Stdout)
		go exec.EnableOutput(nil, r.remote.Stderr)
	}

	// Install containerd if service is not running.
	if err := r.runR("get-service -name containerd"); err != nil {
		cmd := fmt.Sprintf(".\\Install-Containerd.ps1 -ContainerDVersion %s", containerd)
		return r.runR(`curl.exe -LO https://raw.githubusercontent.com/kubernetes-sigs/sig-windows-tools/master/hostprocess/Install-Containerd.ps1; ` + cmd)
	} else if strings.Contains(output, "Running") {
		klog.Info(resc.Sprintf("Skipping containerd installation, service already running, use the copy command."))
	}
	return nil
}

// InstallKubernetes install all Kubernetes bits with the set version.
func (r *Runner) InstallKubernetes(kubernetes string) error {
	klog.Info(mainc.Sprintf("Installing Kubelet."))

	var output string
	if r.Logging {
		go exec.EnableOutput(nil, r.remote.Stdout)
		go exec.EnableOutput(nil, r.remote.Stderr)
	}

	if err := r.runR("get-service -name kubelet"); err != nil {
		// Install Kubernetes if service is not running.
		cmd := fmt.Sprintf(".\\PrepareNode.ps1 -KubernetesVersion %s", kubernetes)
		return r.runR(`curl.exe -LO https://raw.githubusercontent.com/kubernetes-sigs/sig-windows-tools/master/hostprocess/PrepareNode.ps1; ` + cmd)
	} else if strings.Contains(output, "kubelet") { // Otherwise skip
		klog.Info(resc.Sprintf("Skipping Kubelet installation, service already running, use the copy command."))
	}
	return nil
}

// JoinNode joins the Windows node into control-plane cluster.
func (r *Runner) JoinNode(cpVersion, cpIPAddr string) error {
	var (
		err             error
		output, loutput string
	)

	go exec.EnableOutput(&output, r.remote.Stdout)
	go exec.EnableOutput(&output, r.remote.Stderr)
	go exec.EnableOutput(&loutput, r.local.Stdout)

	// In case kubelet is already running, skip joining procedure.
	if err = r.runR("get-service -name kubelet"); err == nil && !strings.Contains(output, "Running") {
		fmt.Println("outputttt ", output)
		// Control plane token create and extract, saving the final command
		lcmd := strings.Join([]string{
			"minikube", "ssh", "--", "sudo", fmt.Sprintf("/var/lib/minikube/binaries/%s/kubeadm", cpVersion),
			"token", "create", "--print-join-command",
		}, " ")
		if err = r.runL(lcmd); err != nil {
			return err
		}

		// Force the creation of the minikube folder for certificates
		if err = r.runR("mkdir c:\\var\\lib\\minikube\\certs -Force"); err != nil {
			return err
		}
		// Copy the control plane host value to Windows hosts
		if err = r.runR(fmt.Sprintf(`Add-content -Path C:\\Windows\\System32\\drivers\\etc\\hosts -Value \"%s %s\"`, cpIPAddr, cpHost)); err != nil {
			return err
		}

		// Trigger a goroutine to copy the ca.crt from kubernetes/pki to the CA folder.
		go func() {
		loop:
			for {
				select {
				case <-time.After(1 * time.Second):
					klog.Info(resc.Sprintf("trying to copy cert..."))
					cmd := "cp c:\\etc\\kubernetes\\pki\\ca.crt c:\\var\\lib\\minikube\\certs\\ca.crt"
					if err = r.runRstd(cmd, nil); err == nil {
						break loop
					}
				}
			}
		}()

		// Add the control plane into hosts and start the join command.
		return r.runR(fmt.Sprintf(`$env:Path += ';c:\\k\\'; %s`, strings.Trim(loutput, " \n")))
	}

	klog.Info(resc.Sprintf("Skipping node join, the Kubelet service is already running."))
	return nil
}

// InstallCNI installs Calico CNI receiving a specific version.
func (r *Runner) InstallCNI(calicoVersion, controlPlaneIP string) error {
	var loutput string

	go exec.EnableOutput(&loutput, r.local.Stdout)
	go exec.EnableOutput(&loutput, r.local.Stderr)

	var (
		err     error
		content []byte
	)

	if content, err = templates.OpenYAMLFile("./specs/kube-proxy.yml"); err != nil {
		return err
	}
	kpTmpl := templates.KubeProxyTmpl{KUBERNETES_VERSION: calicoVersion}
	t, _ := templates.ChangeTemplate(string(content), kpTmpl)
	kpTempFile := templates.SaveFile(t)
	defer templates.DeleteFile(kpTempFile)

	if content, err = templates.OpenYAMLFile("./specs/configmap.yml"); err != nil {
		return err
	}
	cpTmpl := templates.ConfigMapTmpl{KUBERNETES_SERVICE_HOST: controlPlaneIP, KUBERNETES_SERVICE_PORT: "8443"}
	t, _ = templates.ChangeTemplate(string(content), cpTmpl)
	cpTempFile := templates.SaveFile(t)
	defer templates.DeleteFile(cpTempFile)

	// Execute Kubernetes steps for Calico installation
	steps := [][]string{
		{"kubectl", "config", "set-context", "minikube"},
		{"kubectl", "create", "-f", fmt.Sprintf("https://raw.githubusercontent.com/projectcalico/calico/%v/manifests/tigera-operator.yaml", calicoVersion)},
		{"kubectl", "create", "-f", "./specs/installation.yaml"},
		{"kubectl", "create", "-f", cpTempFile},
		{"kubectl", "create", "-f", kpTempFile},
		{"kubectl", "patch", "ipamconfigurations", "default", "--type", "merge", "--patch=" + string(templates.GetSpecAffinity())},
	}
	for i := 0; i <= len(steps)-1; i++ {
		cmd := strings.Join(steps[i], " ")
		resc.Printf("Running: %v\n", cmd)
		if err := r.runL(cmd); err != nil {
			bad.Printf("%v", err)
		}
	}

	return nil
}
