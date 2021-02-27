// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cluster

import (
	"context"

	"github.com/bborbe/world/configuration/service"
	"github.com/bborbe/world/pkg/dns"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/network"
	"github.com/bborbe/world/pkg/remote"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Sun struct {
	SSH         *ssh.SSH
	Context     k8s.Context
	ClusterIP   network.IP
	DisableRBAC bool
}

func (s *Sun) Children() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguraionBuilder().WithApplier(
			&dns.Server{
				Host:    "ns.rocketsource.de",
				KeyPath: "/Users/bborbe/.dns/home.benjamin-borbe.de.key",
				List: []dns.Entry{
					{
						Host: "backup.sun.pn.benjamin-borbe.de",
						IP:   s.ClusterIP,
					},
				},
			},
		),
		world.NewConfiguraionBuilder().WithApplier(&remote.IptablesAllowInput{
			SSH:  s.SSH,
			Port: network.PortStatic(80),
		}),
		&service.UbuntuUnattendedUpgrades{
			SSH: s.SSH,
		},
		&service.Kubernetes{
			SSH:         s.SSH,
			Context:     s.Context,
			ClusterIP:   s.ClusterIP,
			DisableRBAC: s.DisableRBAC,
		},
	}
}

func (s *Sun) Applier() (world.Applier, error) {
	return nil, nil
}

func (s *Sun) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		s.Context,
		s.ClusterIP,
	)
}
