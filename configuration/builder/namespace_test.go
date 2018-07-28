package builder

import (
	"bytes"

	"github.com/go-yaml/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("NamespaceBuilder", func() {
	It("namespace", func() {
		namespaceBuilder := &NamespaceBuilder{
			Namespace: "banana",
		}
		b := &bytes.Buffer{}
		err := yaml.NewEncoder(b).Encode(namespaceBuilder.Build())
		Expect(err).NotTo(HaveOccurred())
		Expect(b.String()).To(Equal(`apiVersion: v1
kind: Namespace
metadata:
  name: banana
  labels:
    app: banana
`))
	})
})
