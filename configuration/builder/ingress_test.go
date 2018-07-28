package builder

import (
	"bytes"

	"github.com/bborbe/world"
	"github.com/go-yaml/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("IngressBuilder", func() {
	It("ingress", func() {
		ingressBuilder := &IngressBuilder{
			Namespace: "banana",
			Domains:   []world.Domain{"example.com"},
		}
		b := &bytes.Buffer{}
		err := yaml.NewEncoder(b).Encode(ingressBuilder.Build())
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
          servicePort: web
        path: /
`))
	})
})
