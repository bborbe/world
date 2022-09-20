// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service

import (
	"context"

	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Ubuntu struct {
	SSH *ssh.SSH
}

func (u *Ubuntu) Children(ctx context.Context) (world.Configurations, error) {
	return world.Configurations{
		&Smartd{
			SSH: u.SSH,
		},
		&UbuntuUnattendedUpgrades{
			SSH: u.SSH,
		},
		&HdParam{
			SSH: u.SSH,
		},
		&Iptables{
			SSH: u.SSH,
		},
	}, nil
}

func (u *Ubuntu) Applier() (world.Applier, error) {
	return nil, nil
}

func (u *Ubuntu) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		u.SSH,
	)
}
