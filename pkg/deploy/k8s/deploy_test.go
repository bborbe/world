package k8s

import (
	"testing"

	"io/ioutil"

	"github.com/bborbe/world"
	"github.com/bborbe/world/mocks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("K8s Deploy", func() {
	It("hostPort", func() {
		builder := &mocks.Builder{}
		builder.GetImageReturns(world.Image{})
		uploader := &mocks.Uploader{}
		uploader.GetBuilderReturns(builder)
		deployer := &Deployer{
			Uploader: uploader,
			HostPort: 123,
		}
		reader, err := deployer.generateDeployment()
		Expect(err).ToNot(HaveOccurred())
		bytes, err := ioutil.ReadAll(reader)
		Expect(err).ToNot(HaveOccurred())
		Expect(gbytes.BufferWithBytes(bytes)).To(gbytes.Say("hostPort: 123"))

	})
})

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "K8s Deploy Suite")
}
