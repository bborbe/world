// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cluster

import (
	"context"

	service "github.com/bborbe/world/configuration/serivce"
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

func (h *Netcup) Children() []world.Configuration {
	user := ssh.User("bborbe")
	return []world.Configuration{
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
			Version:     h.KubernetesVersion,
		},
	}
}

func (r *Netcup) Applier() (world.Applier, error) {
	return nil, nil
}

func (r *Netcup) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
	)
}
