package kubernetes

import (
	"fmt"
	"github.com/fatih/color"
	klog "k8s.io/klog/v2"
	"swdt/apis/config/v1alpha1"
	"swdt/pkg/executors/iface"
)

var (
	resc       = color.New(color.FgHiGreen).Add(color.Bold)
	permission = "0755"
)

type Runner struct {
	remote iface.SSHExecutor
	local  iface.Executor
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

func (r *Runner) InstallProvisioners(provisioners []v1alpha1.ProvisionerSpec) error {
	for _, provisioner := range provisioners {
		source, destination := provisioner.SourceURL, provisioner.Destination
		name := provisioner.Name
		klog.Info(resc.Sprintf("Service %s binary replacement, trying to stop service...", name))
		err := r.runR(fmt.Sprintf("Stop-Service -name %s -Force", name))
		if err != nil {
			klog.Error(err)
			continue
		}
		klog.Infof("Service stopped. Copying file %s to remote %s...", source, destination)
		/*if err = (*r.remote).Copy(source, destination, permission); err != nil {
			klog.Error(err)
			continue
		}*/
		klog.Infof("starting service %s again...", name)
		err = r.runR(fmt.Sprintf("Start-Service -name %s", name))
		if err != nil {
			klog.Error(err)
			continue
		}
		klog.Info(resc.Sprintf("Service started.\n"))
	}
	return nil
}
