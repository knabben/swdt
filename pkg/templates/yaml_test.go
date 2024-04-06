package templates

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var filepath = "./testdata/kube-proxy.yml"

func TestOpenYAML(t *testing.T) {
	content, err := OpenYAMLFile(filepath)
	assert.Nil(t, err)
	assert.Greater(t, len(content), 0)
	assert.Contains(t, string(content), "{{.KUBERNETES_VERSION}}")
}

func TestRenderTemplate(t *testing.T) {
	content, err := OpenYAMLFile(filepath)
	assert.Nil(t, err)

	var version string = "v1.19.0"
	kpTmpl := KubeProxyTmpl{KUBERNETES_VERSION: version}
	output, err := ChangeTemplate(string(content), kpTmpl)
	assert.Nil(t, err)
	assert.Contains(t, output, version)
}
