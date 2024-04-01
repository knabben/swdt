package executor

import (
	"github.com/stretchr/testify/assert"
	"swdt/apis/config/v1alpha1"
	"swdt/pkg/connections/tests"
	"swdt/pkg/pwsh/kubernetes"
	"swdt/pkg/pwsh/setup"
	"testing"
)

func TestMultipleExecutors(t *testing.T) {
	hostname := tests.GetHostname("2024")
	tests.NewServer(hostname, "fake runner")

	kubeRunner, err := NewRunner(createNodeConfig(hostname), &kubernetes.KubernetesRunner{})
	assert.Nil(t, err)
	assert.NotNil(t, kubeRunner.Inner)
	assert.IsType(t, kubernetes.KubernetesRunner{}, *kubeRunner.Inner)
	err = kubeRunner.CloseConnection()
	assert.Nil(t, err)

	setupRunner, err := NewRunner(createNodeConfig(hostname), &setup.SetupRunner{})
	assert.Nil(t, err)
	assert.NotNil(t, setupRunner.Inner)
	assert.IsType(t, setup.SetupRunner{}, *setupRunner.Inner)
	err = setupRunner.CloseConnection()
	assert.Nil(t, err)
}

func createNodeConfig(hostname string) *v1alpha1.Cluster {
	credentials := v1alpha1.SSHSpec{
		Hostname: hostname,
		Username: tests.Username,
		Password: tests.FakePassword,
	}
	return &v1alpha1.Cluster{Spec: v1alpha1.ClusterSpec{
		Workload: v1alpha1.WorkloadSpec{
			Virtualization: v1alpha1.VirtualizationSpec{
				SSH: &credentials,
			},
		},
	}}
}
