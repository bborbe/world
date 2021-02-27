// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cluster

import (
	"context"

	"github.com/bborbe/world/configuration/service"
	"github.com/bborbe/world/pkg/dns"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/network"
	"github.com/bborbe/world/pkg/remote"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Nuke struct {
	SSH         *ssh.SSH
	Context     k8s.Context
	ClusterIP   network.IP
	DisableRBAC bool
	Version     docker.Tag
}

func (n *Nuke) Children() []world.Configuration {

	return []world.Configuration{
		world.NewConfiguraionBuilder().WithApplier(
			&dns.Server{
				Host:    "ns.rocketsource.de",
				KeyPath: "/Users/bborbe/.dns/home.benjamin-borbe.de.key",
				List: []dns.Entry{
					{
						Host: "backup.nuke.hm.benjamin-borbe.de",
						IP:   n.ClusterIP,
					},
				},
			},
		),
		world.NewConfiguraionBuilder().WithApplier(&remote.IptablesAllowInput{
			SSH:  n.SSH,
			Port: network.PortStatic(80),
		}),
		&service.UbuntuUnattendedUpgrades{
			SSH: n.SSH,
		},
		&service.Kubernetes{
			SSH:         n.SSH,
			Context:     n.Context,
			ClusterIP:   n.ClusterIP,
			DisableRBAC: n.DisableRBAC,
			Version:     n.Version,
		},
	}
}

func (n *Nuke) Applier() (world.Applier, error) {
	return nil, nil
}

func (n *Nuke) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		n.Context,
		n.ClusterIP,
	)
}
