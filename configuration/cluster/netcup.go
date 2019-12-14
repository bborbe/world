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
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Netcup struct {
	Context           k8s.Context
	IP                dns.IP
	DisableRBAC       bool
	DisableCNI        bool
	KubernetesVersion docker.Tag
}

func (n *Netcup) Children() []world.Configuration {
	ssh := &ssh.SSH{
		Host: ssh.Host{
			IP:   n.IP,
			Port: 22,
		},
		User:           ssh.User("bborbe"),
		PrivateKeyPath: "/Users/bborbe/.ssh/id_rsa",
	}
	return []world.Configuration{
		&service.DisablePostfix{
			SSH: ssh,
		},
		&service.SecurityUpdates{
			SSH: ssh,
		},
		&service.Kubernetes{
			SSH:         ssh,
			Context:     n.Context,
			ClusterIP:   n.IP,
			DisableRBAC: n.DisableRBAC,
			DisableCNI:  n.DisableCNI,
			Version:     n.KubernetesVersion,
		},
	}
}

func (n *Netcup) Applier() (world.Applier, error) {
	return nil, nil
}

func (n *Netcup) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
	)
}
