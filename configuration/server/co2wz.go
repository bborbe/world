// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"context"

	"github.com/bborbe/world/pkg/dns"
	"github.com/bborbe/world/pkg/network"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Co2wz struct {
	IP network.IP
}

func (c *Co2wz) Children(ctx context.Context) (world.Configurations, error) {
	return world.Configurations{
		world.NewConfiguraionBuilder().WithApplier(
			&dns.Server{
				Host:    "ns.rocketsource.de",
				KeyPath: "/Users/bborbe/.dns/home.benjamin-borbe.de.key",
				List: []dns.Entry{
					{
						Host: "co2wz.hm.benjamin-borbe.de",
						IP:   c.IP,
					},
				},
			},
		),
	}, nil
}

func (c *Co2wz) Applier() (world.Applier, error) {
	return nil, nil
}

func (c *Co2wz) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
	)
}
