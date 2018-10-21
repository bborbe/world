package k8s

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

func YamlToStruct(reader io.Reader, writer io.Writer) error {
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return errors.Wrap(err, "reader failed")
	}
	data := make(map[string]interface{})
	if err := yaml.NewDecoder(bytes.NewBuffer(content)).Decode(data); err != nil {
		return errors.Wrap(err, "decode yaml failed")
	}
	kind, ok := data["kind"]
	if !ok {
		return errors.New("kind missing in yaml")
	}
	mapping := map[string]interface{}{
		"ConfigMap":           &ConfigMap{},
		"DaemonSet":           &DaemonSet{},
		"Deployment":          &Deployment{},
		"StatefulSet":         &StatefulSet{},
		"Secret":              &Secret{},
		"Namespace":           &Namespace{},
		"Ingress":             &Ingress{},
		"Service":             &Service{},
		"PodDisruptionBudget": &PodDisruptionBudget{},
		"ClusterRole":         &ClusterRole{},
		"ClusterRoleBinding":  &ClusterRoleBinding{},
	}

	s, ok := mapping[kind.(string)]
	if !ok {
		return fmt.Errorf("unkown kind %s", kind)
	}
	if err := yaml.NewDecoder(bytes.NewBuffer(content)).Decode(s); err != nil {
		return errors.Wrapf(err, "decode yaml into %s failed", kind)
	}
	fmt.Fprintf(writer, "%#v\n", s)
	return nil
}
