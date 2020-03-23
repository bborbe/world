// Copyright (c) 2019 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package secret_test

import (
	"context"

	"github.com/bborbe/teamvault-utils"
	"github.com/bborbe/teamvault-utils/mocks"
	"github.com/bborbe/world/pkg/deployer"
	"github.com/bborbe/world/pkg/secret"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Secret", func() {
	var teamvaultSecret *secret.Teamvault
	var teamvaultConnector *mocks.Connector
	var ctx context.Context
	BeforeEach(func() {
		ctx = context.Background()
		teamvaultConnector = &mocks.Connector{}
		teamvaultSecret = &secret.Teamvault{
			TeamvaultConnector: teamvaultConnector,
		}

	})
	Context("Password", func() {
		var password deployer.SecretValue
		BeforeEach(func() {
			password = teamvaultSecret.Password("123")
			teamvaultConnector.PasswordReturns(teamvault.Password("s3cr3t"), nil)
		})
		It("validates successful", func() {
			err := password.Validate(ctx)
			Expect(err).To(BeNil())
		})
		It("calls teamvault connector", func() {
			_, _ = password.Value(ctx)
			Expect(teamvaultConnector.PasswordCallCount()).To(Equal(1))
		})
		It("returns value", func() {
			value, err := password.Value(ctx)
			Expect(err).To(BeNil())
			Expect(string(value)).To(Equal("s3cr3t"))
		})
	})
})
