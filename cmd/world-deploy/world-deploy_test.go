package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("World Deploy", func() {
	It("compiles", func() {
		_, err := gexec.Build("github.com/bborbe/world/cmd/world-deploy")
		Expect(err).NotTo(HaveOccurred())
	})
})

func TestWorldDockerBuild(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "World Deploy Suite")
}
