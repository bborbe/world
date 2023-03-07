// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package deployer

import (
	"bytes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
)

var _ = Describe("ServiceDeployer", func() {
	It("service", func() {
		serviceDeployer := &ServiceDeployer{
			Namespace: "banana",
			Name:      "banana",
			Ports: []Port{
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
