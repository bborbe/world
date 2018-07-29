package deployer

import (
	"bytes"

	"github.com/go-yaml/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ServiceDeployer", func() {
	It("service", func() {
		serviceDeployer := &ServiceDeployer{
			Port:      1337,
			Namespace: "banana",
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
  - name: web
    port: 1337
    protocol: TCP
    targetPort: http
  selector:
    app: banana
`))
	})
})
