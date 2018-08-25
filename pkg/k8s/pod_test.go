package k8s_test

import (
	"testing"

	"bytes"

	"github.com/bborbe/world/pkg/k8s"
	"github.com/go-yaml/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
)

var _ = Describe("Pod", func() {
	format.TruncatedDiff = false
	It("encode empty volume", func() {
		p := k8s.PodSpec{
			Volumes: []k8s.PodVolume{
				{
					Name:     "foo",
					EmptyDir: &k8s.PodVolumeEmptyDir{},
				},
			},
		}
		b := &bytes.Buffer{}
		err := yaml.NewEncoder(b).Encode(p)
		Expect(err).To(BeNil())
		Expect(b.String()).To(Equal(`containers: []
volumes:
- name: foo
  emptyDir: {}
`))
	})
	It("encode nfs volume", func() {
		p := k8s.PodSpec{
			Volumes: []k8s.PodVolume{
				{
					Name: "foo",
					Nfs: k8s.PodVolumeNfs{
						Path:   "/path",
						Server: "127.0.0.1",
					},
				},
			},
		}
		b := &bytes.Buffer{}
		err := yaml.NewEncoder(b).Encode(p)
		Expect(err).To(BeNil())
		Expect(b.String()).To(Equal(`containers: []
volumes:
- name: foo
  nfs:
    path: /path
    server: 127.0.0.1
`))
	})
	It("encode configmap volume", func() {
		p := k8s.PodSpec{
			Volumes: []k8s.PodVolume{
				{
					Name: "foo",
					ConfigMap: k8s.PodVolumeConfigMap{
						Name: "source",
						Items: []k8s.PodConfigMapItem{
							{
								Key:  "config",
								Path: "bar.toml",
							},
						},
					},
				},
			},
		}
		b := &bytes.Buffer{}
		err := yaml.NewEncoder(b).Encode(p)
		Expect(err).To(BeNil())
		Expect(b.String()).To(Equal(`containers: []
volumes:
- name: foo
  configMap:
    name: source
    items:
    - key: config
      path: bar.toml
`))
	})
	It("encode resources", func() {
		p := k8s.Resources{
			Limits: k8s.ContainerResource{
				Cpu:    "1",
				Memory: "2",
			},
			Requests: k8s.ContainerResource{
				Cpu:    "3",
				Memory: "4",
			},
		}
		b := &bytes.Buffer{}
		err := yaml.NewEncoder(b).Encode(p)
		Expect(err).To(BeNil())
		Expect(b.String()).To(Equal(`limits:
  cpu: "1"
  memory: "2"
requests:
  cpu: "3"
  memory: "4"
`))
	})
})

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "K8s Suite")
}
