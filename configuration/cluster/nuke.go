// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cluster

import (
	"context"

	"github.com/bborbe/world/configuration/serivce"
	"github.com/bborbe/world/pkg/dns"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/remote"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Nuke struct {
	Context     k8s.Context
	ClusterIP   dns.IP
	DisableRBAC bool
	DisableCNI  bool
}

func (n *Nuke) Children() []world.Configuration {
	ssh := &ssh.SSH{
		Host: ssh.Host{
			IP:   n.ClusterIP,
			Port: 22,
		},
		User:           "bborbe",
		PrivateKeyPath: "/Users/bborbe/.ssh/id_rsa",
	}
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
		world.NewConfiguraionBuilder().WithApplier(&remote.Iptables{
			SSH:  ssh,
			Port: 80,
		}),
		&service.Kubernetes{
			SSH:         ssh,
			Context:     n.Context,
			ClusterIP:   n.ClusterIP,
			DisableRBAC: n.DisableRBAC,
			DisableCNI:  n.DisableCNI,
			ResolvConf:  "/run/systemd/resolve/resolv.conf",
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
