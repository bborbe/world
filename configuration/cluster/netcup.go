// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cluster

import (
	"context"

	"github.com/bborbe/world/configuration/service"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/network"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Netcup struct {
	Context           k8s.Context
	IP                network.IP
	DisableRBAC       bool
	KubernetesVersion docker.Tag
	SSH               *ssh.SSH
}

func (n *Netcup) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		n.SSH,
	)
}

func (n *Netcup) Children() []world.Configuration {
	return []world.Configuration{
		&service.DisablePostfix{
			SSH: n.SSH,
		},
		&service.UbuntuUnattendedUpgrades{
			SSH: n.SSH,
		},
		&service.Kubernetes{
			SSH:         n.SSH,
			Context:     n.Context,
			ClusterIP:   n.IP,
			DisableRBAC: n.DisableRBAC,
			Version:     n.KubernetesVersion,
		},
	}
}

func (n *Netcup) Applier() (world.Applier, error) {
	return nil, nil
}
