package templates

import (
	"bytes"
	"encoding/json"
	"io"
	"os"

	"text/template"
)

type KubeProxyTmpl struct {
	KUBERNETES_VERSION string
}

type ConfigMapTmpl struct {
	KUBERNETES_SERVICE_HOST string
	KUBERNETES_SERVICE_PORT string
}

type SpecData struct {
	Spec struct {
		StrictAffinity bool `json:"strictAffinity,omitempty"`
	} `json:"spec,omitempty"`
}

// ChangeTemplate overwrite the pre-defined text template based in the input struct
func ChangeTemplate[T KubeProxyTmpl | ConfigMapTmpl](mapping string, tmplStruct T) (string, error) {
	var result bytes.Buffer
	// Parse template and apply changes from the struct
	tmpl := template.Must(template.New("render").Parse(mapping))
	if err := tmpl.Execute(&result, tmplStruct); err != nil {
		return "", err
	}
	// Returns the resulted rendering.
	return result.String(), nil
}

// OpenYAMLFile renders the YAML file and returns its content
func OpenYAMLFile(filename string) (content []byte, err error) {
	var fd *os.File
	if fd, err = os.Open(filename); err != nil {
		return
	}
	content, err = io.ReadAll(fd)
	return
}

// GetSpecAffinity render affinity struct for Calico patch
func GetSpecAffinity() (content []byte) {
	var data = SpecData{}
	data.Spec.StrictAffinity = true
	content, _ = json.Marshal(data)
	return
}
