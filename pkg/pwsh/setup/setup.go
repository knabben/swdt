package setup

import (
	"fmt"
	"github.com/fatih/color"
	klog "k8s.io/klog/v2"
	"strings"
	"swdt/pkg/connections"
)

var (
	mainc = color.New(color.FgHiBlack).Add(color.Underline)
	resc  = color.New(color.FgHiGreen).Add(color.Bold)
)

const (
	CHOCO_PATH    = "C:\\ProgramData\\chocolatey\\bin\\choco.exe"
	CHOCO_INSTALL = "install --accept-licenses --yes"
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
	if err != nil {
		return err
	}
	klog.Info(resc.Sprintf("Installed Choco with: %s", output))

	return nil
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
	if err != nil {
		return err
	}
	return nil
}

func (r *SetupRunner) InstallContainerd(containerd string) error {
	var (
		err    error
		output string
	)
	// Install containerd if service is not running.
	if output, err = r.run("get-service -name containerd"); err != nil {
		cmd := fmt.Sprintf(".\\Install-Containerd.ps1 -ContainerDVersion %s", containerd)
		if output, err = r.run(
			`curl.exe -LO https://raw.githubusercontent.com/kubernetes-sigs/sig-windows-tools/master/hostprocess/Install-Containerd.ps1; ` + cmd,
		); err != nil {
			return err
		}
		klog.Info(resc.Sprintf("%s -- Containerd v%s installed with sucess.", output, containerd))
		return nil
	} else if strings.Contains(output, "Running") {
		// Otherwise skip
		klog.Info(resc.Sprintf("Skipping containerd installation, service already running."))
		return nil
	}
	return nil
}
