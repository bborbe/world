// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cluster

import (
	"context"

	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/configuration/service"
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
}

func (h *Hetzner) Children() []world.Configuration {
	user := ssh.User("bborbe")
	ssh := &ssh.SSH{
		Host: ssh.Host{
			IP:   h.IP,
			Port: 22,
		},
		User:           user,
		PrivateKeyPath: "/Users/bborbe/.ssh/id_rsa",
	}
	return []world.Configuration{
		//world.NewConfiguraionBuilder().WithApplier(&hetzner.Server{
		//	ApiKey:        h.ApiKey,
		//	Name:          h.Context,
		//	User:          user,
		//	PublicKeyPath: "/Users/bborbe/.ssh/id_rsa.pub",
		//	ServerType:    h.ServerType,
		//}),
		//world.NewConfiguraionBuilder().WithApplier(
		//	&dns.Server{
		//		Host:    "ns.rocketsource.de",
		//		KeyPath: "/Users/bborbe/.dns/home.benjamin-borbe.de.key",
		//		List: []dns.Entry{
		//			{
		//				Host: dns.Host(fmt.Sprintf("%s.benjamin-borbe.de", h.Context.String())),
		//				IP:   h.IP,
		//			},
		//		},
		//	},
		//),
		//&service.UbuntuUnattendedUpgrades{
		//	SSH: ssh,
		//},
		//&service.Kubernetes{
		//	SSH:         ssh,
		//	Context:     h.Context,
		//	ClusterIP:   h.IP,
		//	DisableRBAC: h.DisableRBAC,
		//	DisableCNI:  h.DisableCNI,
		//	Version:     h.KubernetesVersion,
		//},
		&service.OpenvpnServer{
			ServerName: "hetzner",
			SSH:        ssh,
			ServerIP:   "192.168.0.0",
			ServerMask: "255.255.255.0",
			Routes:     []service.OpenvpnRoute{},
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
