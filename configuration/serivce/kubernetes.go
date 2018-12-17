// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service

import (
	"context"

	"github.com/bborbe/world/pkg/dns"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Kubernetes struct {
	SSH         *ssh.SSH
	Context     k8s.Context
	ClusterIP   dns.IP
	DisableRBAC bool
	DisableCNI  bool
	ResolvConf  string
}

func (k *Kubernetes) Children() []world.Configuration {
	return []world.Configuration{
		&Etcd{
			SSH: k.SSH,
		},
		&Kubelet{
			SSH:         k.SSH,
			Version:     "v1.11.5",
			Context:     k.Context,
			ClusterIP:   k.ClusterIP,
			DisableRBAC: k.DisableRBAC,
			DisableCNI:  k.DisableCNI,
			ResolvConf:  k.ResolvConf,
		},
	}
}

func (k *Kubernetes) Applier() (world.Applier, error) {
	return nil, nil
}

func (k *Kubernetes) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		k.SSH,
		k.Context,
	)
}
