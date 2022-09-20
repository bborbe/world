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

type Fire struct {
	IP  network.IP
	SSH *ssh.SSH
}

func (f *Fire) Children(ctx context.Context) (world.Configurations, error) {
	return world.Configurations{
		world.NewConfiguraionBuilder().WithApplier(
			&dns.Server{
				Host:    "ns.rocketsource.de",
				KeyPath: "/Users/bborbe/.dns/home.benjamin-borbe.de.key",
				List: []dns.Entry{
					{
						Host: "fire.hm.benjamin-borbe.de",
						IP:   f.IP,
					},
				},
			},
		),
		world.NewConfiguraionBuilder().WithApplier(&remote.IptablesAllowInput{
			SSH:      f.SSH,
			Port:     network.PortStatic(80),
			Protocol: network.TCP,
		}),
		&service.Ubuntu{
			SSH: f.SSH,
		},
	}, nil
}

func (f *Fire) Applier() (world.Applier, error) {
	return nil, nil
}

func (f *Fire) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		f.IP,
	)
}
