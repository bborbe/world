// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package k8s_test

import (
	"bytes"
	"testing"

	"github.com/bborbe/world/pkg/k8s"
	"github.com/go-yaml/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
)

func TestPodDisruptionBudget(t *testing.T) {

}

var _ = Describe("PodDisruptionBudget", func() {
	format.TruncatedDiff = true
	It("encode", func() {
		content := `apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  name: test-pdb
  labels:
    app: test
spec:
  maxUnavailable: 1
  minAvailable: 1
  selector:
    matchLabels:
      app: test
`
		var data k8s.PodDisruptionBudget
		err := yaml.NewDecoder(bytes.NewBufferString(content)).Decode(&data)
		Expect(err).To(BeNil())

		var b bytes.Buffer
		err = yaml.NewEncoder(&b).Encode(data)
		Expect(err).To(BeNil())

		Expect(content).To(Equal(b.String()))
	})
})
