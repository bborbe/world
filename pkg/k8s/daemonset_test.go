package k8s_test

import (
	"bytes"

	"github.com/bborbe/world/pkg/k8s"
	"github.com/go-yaml/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
)

var _ = Describe("DaemonSet", func() {
	format.TruncatedDiff = false
	It("encode simple daemonSet", func() {
		p := k8s.DaemonSet{
			ApiVersion: "v1",
			Kind:       "DaemonSet",
			Metadata: k8s.Metadata{
				Name:      "d",
				Namespace: "n",
			},
			Spec: k8s.DaemonSetSpec{
				Template: k8s.PodTemplate{
					Metadata: k8s.Metadata{
						Annotations: k8s.Annotations{
							"a": "b",
						},
						Labels: k8s.Labels{
							"c": "d",
						},
					},
					Spec: k8s.PodSpec{
						Containers: []k8s.Container{
							{
								Name:  "c",
								Image: "i",
							},
						},
					},
				},
			},
		}
		b := &bytes.Buffer{}
		err := yaml.NewEncoder(b).Encode(p)
		Expect(err).To(BeNil())
		Expect(b.String()).To(Equal(`apiVersion: v1
kind: DaemonSet
metadata:
  namespace: "n"
  name: d
spec:
  template:
    metadata:
      labels:
        c: d
      annotations:
        a: b
    spec:
      containers:
      - name: c
        image: i
`))
	})
})