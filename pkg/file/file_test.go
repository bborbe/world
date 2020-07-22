// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package file_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/bborbe/world/pkg/file"
)

var _ = Describe("Remote", func() {
	It("Perm 755 to string", func() {
		var perm file.Perm = 0755
		Expect(perm.String()).To(Equal("0755"))
	})
	It("Perm 644 to string", func() {
		var perm file.Perm = 0644
		Expect(perm.String()).To(Equal("0644"))
	})
})
