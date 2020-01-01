// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package deployer

import (
	"bytes"
	"context"

	"github.com/bborbe/world/pkg/k8s"
	"github.com/go-yaml/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SecretApplier", func() {
	var secretDeployer *SecretApplier
	var secret *k8s.Secret
	var err error
	BeforeEach(func() {
		secretDeployer = &SecretApplier{
			Namespace: "banana",
			Name:      "banana",
			Secrets: Secrets{
				"secret": SecretValueStatic([]byte("hello world")),
			},
		}
		secret, err = secretDeployer.secret(context.Background())
	})
	It("returns no error", func() {
		Expect(err).To(BeNil())
	})
	It("secret", func() {
		Expect(secret).NotTo(BeNil())
		b := &bytes.Buffer{}
		err = yaml.NewEncoder(b).Encode(secret)
		Expect(err).To(BeNil())
		Expect(b.String()).To(Equal(`apiVersion: v1
kind: Secret
metadata:
  namespace: banana
  name: banana
  labels:
    app: banana
type: Opaque
data:
  secret: aGVsbG8gd29ybGQ=
`))
	})
})
