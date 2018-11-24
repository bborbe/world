// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cluster

import (
	"context"
	"fmt"

	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/configuration/serivce"
	"github.com/bborbe/world/pkg/dns"
	"github.com/bborbe/world/pkg/hetzner"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Hetzner struct {
	Context     k8s.Context
	ApiKey      deployer.SecretValue
	IP          dns.IP
	DisableRBAC bool
	DisableCNI  bool
}

func (h *Hetzner) Children() []world.Configuration {
	user := ssh.User("bborbe")
	return []world.Configuration{
		world.NewConfiguraionBuilder().WithApplier(&hetzner.Server{
			ApiKey:        h.ApiKey,
			Name:          h.Context,
			User:          user,
			PublicKeyPath: "/Users/bborbe/.ssh/id_rsa.pub",
		}),
		world.NewConfiguraionBuilder().WithApplier(
			&dns.Server{
				Host:    "ns.rocketsource.de",
				KeyPath: "/Users/bborbe/.dns/home.benjamin-borbe.de.key",
				List: []dns.Entry{
					{
						Host: dns.Host(fmt.Sprintf("%s.benjamin-borbe.de", h.Context.String())),
						IP:   h.IP,
					},
				},
			},
		),
		&service.Kubernetes{
			SSH: &ssh.SSH{
				Host: ssh.Host{
					IP:   h.IP,
					Port: 22,
				},
				User:           user,
				PrivateKeyPath: "/Users/bborbe/.ssh/id_rsa",
			},
			Context:     h.Context,
			ClusterIP:   h.IP,
			DisableRBAC: h.DisableRBAC,
			DisableCNI:  h.DisableCNI,
			ResolvConf:  "/run/systemd/resolve/resolv.conf",
		},
	}
}

func (h *Hetzner) Applier() (world.Applier, error) {
	return nil, nil
}

func (h *Hetzner) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		h.Context,
		h.ApiKey,
	)
}
