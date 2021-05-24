// Copyright (c) 2020 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/bborbe/world/configuration/service"
)

var _ = Describe("CronName", func() {
	var ctx context.Context
	var err error
	var cronName service.CronName
	BeforeEach(func() {
		ctx = context.Background()
	})
	Context("Validate", func() {
		Context("valid", func() {
			BeforeEach(func() {
				cronName = "valid-cron_1337"
				err = cronName.Validate(ctx)
			})
			It("returns no error", func() {
				Expect(err).To(BeNil())
			})
		})
		Context("invalid", func() {
			BeforeEach(func() {
				cronName = "in.valid"
				err = cronName.Validate(ctx)
			})
			It("returns error", func() {
				Expect(err).NotTo(BeNil())
			})
		})
	})
	Context("BuildCronName", func() {
		It("simple name", func() {
			Expect(service.BuildCronName("name")).To(Equal(service.CronName("name")))
		})
		It("multi parts", func() {
			Expect(service.BuildCronName("a", "b", "c")).To(Equal(service.CronName("a_b_c")))
		})
		It("invalid chars", func() {
			Expect(service.BuildCronName("hello.world")).To(Equal(service.CronName("hello_world")))
		})
		It("lowercase", func() {
			Expect(service.BuildCronName("Name")).To(Equal(service.CronName("name")))
		})
		It("multi underscore", func() {
			Expect(service.BuildCronName("a____b")).To(Equal(service.CronName("a_b")))
		})
	})
})
