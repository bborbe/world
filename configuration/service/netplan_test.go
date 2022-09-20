// Copyright (c) 2021 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service_test

import (
	"context"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/bborbe/world/configuration/service"
	"github.com/bborbe/world/pkg/network"
	"github.com/bborbe/world/pkg/remote"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/world"
)

var _ = Describe("NetPlan", func() {
	var ctx context.Context
	var err error
	var netPlan service.NetPlan
	BeforeEach(func() {
		ctx = context.Background()
		netPlan = service.NetPlan{
			SSH: &ssh.SSH{
				Host: ssh.Host{
					IP:   network.IPStatic("192.168.2.2"),
					Port: 22,
				},
				PrivateKeyPath: "/tmp/my-key",
				User:           "my-user",
			},
			IP:      network.IPStatic("192.168.2.2"),
			Gateway: network.IPStatic("192.168.2.1"),
			IPMask:  network.MaskStatic(24),
			Device:  "enp3s0",
		}
	})
	Context("Validate", func() {
		JustBeforeEach(func() {
			err = netPlan.Validate(ctx)
		})
		It("returns no error", func() {
			Expect(err).To(BeNil())
		})
	})
	Context("Children", func() {
		var children world.Configurations
		JustBeforeEach(func() {
			children, err = netPlan.Children(ctx)
		})
		It("returns no error", func() {
			Expect(err).To(BeNil())
		})
		It("has childen", func() {
			Expect(len(children)).To(BeNumerically(">", 0))
		})
		It("func", func() {
			var file *remote.File
			for _, configuration := range children {
				switch c := configuration.(type) {
				case *remote.File:
					file = c
				}
			}
			Expect(file).NotTo(BeNil())
			content, err := file.Content.Content(ctx)
			Expect(err).To(BeNil())
			Expect(strings.TrimSpace(string(content))).To(Equal(strings.TrimSpace(`
# This is the network config written by 'world'
network:
  version: 2
  ethernets:
    enp3s0:
      addresses: [192.168.2.2/24]
      gateway4: 192.168.2.1
      nameservers:
        addresses: [8.8.4.4, 8.8.8.8]

`)))
		})
	})
})
