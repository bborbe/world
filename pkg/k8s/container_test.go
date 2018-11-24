// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package k8s_test

import (
	"bytes"

	"github.com/bborbe/world/pkg/k8s"
	"github.com/go-yaml/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
)

var _ = Describe("Container", func() {
	format.TruncatedDiff = false
	It("encode simple container", func() {
		p := k8s.Container{
			Name:  "foo",
			Image: "dummy",
		}
		b := &bytes.Buffer{}
		err := yaml.NewEncoder(b).Encode(p)
		Expect(err).To(BeNil())
		Expect(b.String()).To(Equal(`name: foo
image: dummy
`))
	})
	It("encode container with livenessProbe", func() {
		p := k8s.Container{
			Name:  "foo",
			Image: "dummy",
			LivenessProbe: k8s.Probe{
				HttpGet: k8s.HttpGet{
					Path:   "/healthz",
					Port:   8080,
					Scheme: "HTTP",
				},
				InitialDelaySeconds: 60,
				SuccessThreshold:    1,
				FailureThreshold:    5,
				TimeoutSeconds:      10,
			},
		}
		b := &bytes.Buffer{}
		err := yaml.NewEncoder(b).Encode(p)
		Expect(err).To(BeNil())
		Expect(b.String()).To(Equal(`name: foo
image: dummy
livenessProbe:
  httpGet:
    path: /healthz
    port: 8080
    scheme: HTTP
  initialDelaySeconds: 60
  successThreshold: 1
  failureThreshold: 5
  timeoutSeconds: 10
`))
	})
	It("encode container with readinessProbe", func() {
		p := k8s.Container{
			Name:  "foo",
			Image: "dummy",
			ReadinessProbe: k8s.Probe{
				HttpGet: k8s.HttpGet{
					Path:   "/healthz",
					Port:   8080,
					Scheme: "HTTP",
				},
				InitialDelaySeconds: 60,
				SuccessThreshold:    1,
				FailureThreshold:    5,
				TimeoutSeconds:      10,
			},
		}
		b := &bytes.Buffer{}
		err := yaml.NewEncoder(b).Encode(p)
		Expect(err).To(BeNil())
		Expect(b.String()).To(Equal(`name: foo
image: dummy
readinessProbe:
  httpGet:
    path: /healthz
    port: 8080
    scheme: HTTP
  initialDelaySeconds: 60
  successThreshold: 1
  failureThreshold: 5
  timeoutSeconds: 10
`))
	})
})
