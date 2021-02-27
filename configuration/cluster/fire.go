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

type Fire struct {
	SSH         *ssh.SSH
	Context     k8s.Context
	ClusterIP   network.IP
	DisableRBAC bool
	DisableCNI  bool
}

func (f *Fire) Children() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguraionBuilder().WithApplier(
			&dns.Server{
				Host:    "ns.rocketsource.de",
				KeyPath: "/Users/bborbe/.dns/home.benjamin-borbe.de.key",
				List: []dns.Entry{
					{
						Host: "backup.fire.hm.benjamin-borbe.de",
						IP:   f.ClusterIP,
					},
				},
			},
		),
		world.NewConfiguraionBuilder().WithApplier(&remote.IptablesAllowInput{
			SSH:  f.SSH,
			Port: network.PortStatic(80),
		}),
		&service.UbuntuUnattendedUpgrades{
			SSH: f.SSH,
		},
		&service.Kubernetes{
			SSH:         f.SSH,
			Context:     f.Context,
			ClusterIP:   f.ClusterIP,
			DisableRBAC: f.DisableRBAC,
		},
	}
}

func (f *Fire) Applier() (world.Applier, error) {
	return nil, nil
}

func (f *Fire) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		f.Context,
		f.ClusterIP,
	)
}
