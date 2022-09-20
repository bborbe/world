// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"context"

	"github.com/bborbe/world/configuration/service"
	"github.com/bborbe/world/pkg/dns"
	"github.com/bborbe/world/pkg/network"
	"github.com/bborbe/world/pkg/remote"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Sun struct {
	SSH *ssh.SSH
	IP  network.IP
}

func (s *Sun) Children(ctx context.Context) (world.Configurations, error) {
	return world.Configurations{
		world.NewConfiguraionBuilder().WithApplier(
			&dns.Server{
				Host:    "ns.rocketsource.de",
				KeyPath: "/Users/bborbe/.dns/home.benjamin-borbe.de.key",
				List: []dns.Entry{
					{
						Host: "sun.pn.benjamin-borbe.de",
						IP:   s.IP,
					},
				},
			},
		),
		world.NewConfiguraionBuilder().WithApplier(&remote.IptablesAllowInput{
			SSH:      s.SSH,
			Port:     network.PortStatic(80),
			Protocol: network.TCP,
		}),
		&service.Ubuntu{
			SSH: s.SSH,
		},
	}, nil
}

func (s *Sun) Applier() (world.Applier, error) {
	return nil, nil
}

func (s *Sun) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		s.IP,
	)
}
