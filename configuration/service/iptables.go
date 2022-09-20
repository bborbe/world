// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service

import (
	"context"

	"github.com/bborbe/world/pkg/apt"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Iptables struct {
	SSH *ssh.SSH
}

func (d *Iptables) Children(ctx context.Context) (world.Configurations, error) {
	return world.Configurations{
		world.NewConfiguraionBuilder().WithApplier(&apt.Install{
			SSH:     d.SSH,
			Package: "iptables",
		}),
		world.NewConfiguraionBuilder().WithApplier(&apt.Install{
			SSH:     d.SSH,
			Package: "iptables-persistent",
		}),
		world.NewConfiguraionBuilder().WithApplier(&apt.Install{
			SSH:     d.SSH,
			Package: "netfilter-persistent",
		}),
	}, nil
}

func (d *Iptables) Applier() (world.Applier, error) {
	return nil, nil
}

func (d *Iptables) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		d.SSH,
	)
}
