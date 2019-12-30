// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cluster

import (
	"context"

	"github.com/bborbe/world/pkg/dns"
	"github.com/bborbe/world/pkg/network"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Nova struct {
	IP   network.IP
	Host network.Host
}

func (h *Nova) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		h.Host,
		h.IP,
	)
}

func (h *Nova) Children() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguraionBuilder().WithApplier(
			&dns.Server{
				Host:    "ns.rocketsource.de",
				KeyPath: "/Users/bborbe/.dns/home.benjamin-borbe.de.key",
				List: []dns.Entry{
					{
						Host: h.Host,
						IP:   h.IP,
					},
				},
			},
		),
	}
}

func (h *Nova) Applier() (world.Applier, error) {
	return nil, nil
}
