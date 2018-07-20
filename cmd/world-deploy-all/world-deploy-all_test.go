package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("World Deploy All", func() {
	It("compiles", func() {
		_, err := gexec.Build("github.com/bborbe/world/cmd/world-deploy-all")
		Expect(err).NotTo(HaveOccurred())
	})
})

func TestWorldDockerBuild(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "World Deploy All Suite")
}
