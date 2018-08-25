package deployer

import (
	"bytes"

	"github.com/go-yaml/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("NamespaceDeployer", func() {
	It("namespace", func() {
		namespaceDeployer := &NamespaceDeployer{
			Namespace: "banana",
		}
		b := &bytes.Buffer{}
		err := yaml.NewEncoder(b).Encode(namespaceDeployer.namespace())
		Expect(err).NotTo(HaveOccurred())
		Expect(b.String()).To(Equal(`apiVersion: v1
kind: Namespace
metadata:
  namespace: banana
  name: banana
  labels:
    app: banana
`))
	})
})
