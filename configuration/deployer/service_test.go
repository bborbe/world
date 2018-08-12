package deployer

import (
	"bytes"

	"github.com/bborbe/world"
	"github.com/go-yaml/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ServiceDeployer", func() {
	It("service", func() {
		serviceDeployer := &ServiceDeployer{
			Namespace: "banana",
			Name:      "banana",
			Ports: []world.Port{
				{
					Name: "root",
					Port: 1337,
				},
			},
		}
		b := &bytes.Buffer{}
		err := yaml.NewEncoder(b).Encode(serviceDeployer.service())
		Expect(err).NotTo(HaveOccurred())
		Expect(b.String()).To(Equal(`apiVersion: v1
kind: Service
metadata:
  namespace: banana
  name: banana
  labels:
    app: banana
spec:
  ports:
  - name: root
    port: 1337
  selector:
    app: banana
`))
	})
})
