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

type Hell struct {
	SSH *ssh.SSH
	IP  network.IP
}

func (h *Hell) Children(ctx context.Context) (world.Configurations, error) {
	return world.Configurations{
		world.NewConfiguraionBuilder().WithApplier(
			&dns.Server{
				Host:    "ns.rocketsource.de",
				KeyPath: "/Users/bborbe/.dns/home.benjamin-borbe.de.key",
				List: []dns.Entry{
					{
						Host: "hell.hm.benjamin-borbe.de",
						IP:   h.IP,
					},
				},
			},
		),
		world.NewConfiguraionBuilder().WithApplier(&remote.IptablesAllowInput{
			SSH:      h.SSH,
			Port:     network.PortStatic(22),
			Protocol: network.TCP,
		}),
		&service.Ubuntu{
			SSH: h.SSH,
		},
	}, nil
}

func (h *Hell) Applier() (world.Applier, error) {
	return nil, nil
}

func (h *Hell) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		h.IP,
	)
}
