package deployer

import (
	"bytes"

	"github.com/bborbe/world/pkg/k8s"
	"github.com/go-yaml/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ConfigMapDeployer", func() {
	It("configMap", func() {
		configMapDeployer := &ConfigMapDeployer{
			Namespace: "banana",
			Name:      "banana",
			ConfigMapData: k8s.ConfigMapData{
				"key": "value",
			},
		}
		b := &bytes.Buffer{}
		configMap := configMapDeployer.configMap()
		err := yaml.NewEncoder(b).Encode(configMap)
		Expect(err).NotTo(HaveOccurred())
		Expect(b.String()).To(Equal(`apiVersion: v1
kind: ConfigMap
metadata:
  namespace: banana
  name: banana
  labels:
    app: banana
data:
  key: value
`))
	})
})
