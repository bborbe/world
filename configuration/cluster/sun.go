// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cluster

import (
	"context"

	"github.com/bborbe/world/configuration/service"
	"github.com/bborbe/world/pkg/dns"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/remote"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Sun struct {
	Context     k8s.Context
	ClusterIP   dns.IP
	DisableRBAC bool
	DisableCNI  bool
}

func (s *Sun) Children() []world.Configuration {
	ssh := &ssh.SSH{
		Host: ssh.Host{
			IP:   s.ClusterIP,
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
						Host: "backup.sun.pn.benjamin-borbe.de",
						IP:   s.ClusterIP,
					},
				},
			},
		),
		world.NewConfiguraionBuilder().WithApplier(&remote.Iptables{
			SSH:  ssh,
			Port: 80,
		}),
		&service.SecurityUpdates{
			SSH: ssh,
		},
		&service.Kubernetes{
			SSH:         ssh,
			Context:     s.Context,
			ClusterIP:   s.ClusterIP,
			DisableRBAC: s.DisableRBAC,
			DisableCNI:  s.DisableCNI,
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
