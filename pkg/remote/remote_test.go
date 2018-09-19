package remote_test

import (
	"github.com/bborbe/world/pkg/remote"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Remote", func() {
	It("Perm 755 to string", func() {
		var perm remote.Perm = 0755
		Expect(perm.String()).To(Equal("0755"))
	})
	It("Perm 644 to string", func() {
		var perm remote.Perm = 0644
		Expect(perm.String()).To(Equal("0644"))
	})
})
