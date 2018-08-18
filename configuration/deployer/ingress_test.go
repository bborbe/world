package deployer

import (
	"bytes"

	"github.com/go-yaml/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("IngressDeployer", func() {
	It("ingress", func() {
		ingressDeployer := &IngressDeployer{
			Namespace: "banana",
			Name:      "banana",
			Domains:   []Domain{"example.com"},
		}
		b := &bytes.Buffer{}
		err := yaml.NewEncoder(b).Encode(ingressDeployer.ingress())
		Expect(err).NotTo(HaveOccurred())
		Expect(b.String()).To(Equal(`apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  namespace: banana
  name: banana
  labels:
    app: banana
  annotations:
    kubernetes.io/ingress.class: traefik
    traefik.frontend.priority: "10000"
spec:
  rules:
  - host: example.com
    http:
      paths:
      - backend:
          serviceName: banana
          servicePort: http
        path: /
`))
	})
})
