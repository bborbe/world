package configuration_test

import (
	"testing"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/configuration"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Configuration", func() {
	It("implements world.Configuration interface", func() {
		var i interface{} = configuration.New()
		_, ok := i.(world.Configuration)
		Expect(ok).To(BeTrue())
	})
})

func TestConfiguration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Configuration Suite")
}
