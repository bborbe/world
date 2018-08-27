package k8s_test

import (
	"bytes"

	"github.com/bborbe/world/pkg/k8s"
	"github.com/go-yaml/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
)

var _ = Describe("DeploymentStrategy", func() {
	format.TruncatedDiff = false
	It("encode DeploymentStrategy RollingUpdate", func() {
		p := k8s.DeploymentStrategy{
			Type: "RollingUpdate",
			RollingUpdate: k8s.DeploymentStrategyRollingUpdate{
				MaxSurge:       2,
				MaxUnavailable: 3,
			},
		}
		b := &bytes.Buffer{}
		err := yaml.NewEncoder(b).Encode(p)
		Expect(err).To(BeNil())
		Expect(b.String()).To(Equal(`type: RollingUpdate
rollingUpdate:
  maxSurge: 2
  maxUnavailable: 3
`))
	})
	It("encode DeploymentStrategy Recreate", func() {
		p := k8s.DeploymentStrategy{
			Type: "Recreate",
		}
		b := &bytes.Buffer{}
		err := yaml.NewEncoder(b).Encode(p)
		Expect(err).To(BeNil())
		Expect(b.String()).To(Equal(`type: Recreate
`))
	})
})
