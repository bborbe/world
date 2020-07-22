// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package deployer

import (
	"bytes"

	"github.com/go-yaml/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/bborbe/world/pkg/k8s"
)

var _ = Describe("IngressDeployer", func() {
	It("ingress", func() {
		ingressDeployer := &IngressDeployer{
			Context:   "banana",
			Namespace: "banana",
			Name:      "banana",
			Port:      "http",
			Domains:   k8s.IngressHosts{"example.com"},
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
