package world_test

import (
	"testing"

	"github.com/bborbe/world"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Configuration", func() {
	It("implements world.Configuration interface", func() {
		var i interface{} = world.NewConfiguration()
		_, ok := i.(world.Configuration)
		Expect(ok).To(BeTrue())
	})
})

func TestConfiguration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "World Suite")
}
