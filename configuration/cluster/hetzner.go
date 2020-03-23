// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cluster

import (
	"context"
	"fmt"

	"github.com/bborbe/world/configuration/service"
	"github.com/bborbe/world/pkg/deployer"
	"github.com/bborbe/world/pkg/dns"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/hetzner"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/network"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Hetzner struct {
	Context           k8s.Context
	ApiKey            deployer.SecretValue
	IP                network.IP
	DisableRBAC       bool
	DisableCNI        bool
	KubernetesVersion docker.Tag
	ServerType        hetzner.ServerType
	User              ssh.User
	SSH               *ssh.SSH
}

func (h *Hetzner) Children() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguraionBuilder().WithApplier(&hetzner.Server{
			ApiKey:        h.ApiKey,
			Name:          h.Context,
			User:          h.User,
			PublicKeyPath: "/Users/bborbe/.ssh/id_rsa.pub",
			ServerType:    h.ServerType,
		}),
		world.NewConfiguraionBuilder().WithApplier(
			&dns.Server{
				Host:    "ns.rocketsource.de",
				KeyPath: "/Users/bborbe/.dns/home.benjamin-borbe.de.key",
				List: []dns.Entry{
					{
						Host: network.Host(fmt.Sprintf("%s.benjamin-borbe.de", h.Context.String())),
						IP:   h.IP,
					},
				},
			},
		),
		&service.UbuntuUnattendedUpgrades{
			SSH: h.SSH,
		},
		&service.Kubernetes{
			SSH:         h.SSH,
			Context:     h.Context,
			ClusterIP:   h.IP,
			DisableRBAC: h.DisableRBAC,
			DisableCNI:  h.DisableCNI,
			Version:     h.KubernetesVersion,
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
