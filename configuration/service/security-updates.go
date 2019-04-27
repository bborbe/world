// Copyright (c) 2019 Benjamin Borbe All rights reserved.
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

type SecurityUpdates struct {
	SSH *ssh.SSH
}

func (f *SecurityUpdates) Children() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguraionBuilder().WithApplier(&apt.Update{
			SSH: f.SSH,
		}),
		world.NewConfiguraionBuilder().WithApplier(&apt.Install{
			SSH:     f.SSH,
			Package: "unattended-upgrades",
		}),
		world.NewConfiguraionBuilder().WithApplier(&apt.Autoremove{
			SSH: f.SSH,
		}),
		world.NewConfiguraionBuilder().WithApplier(&apt.Clean{
			SSH: f.SSH,
		}),
	}
}

func (f *SecurityUpdates) Applier() (world.Applier, error) {
	return nil, nil
}

func (f *SecurityUpdates) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		f.SSH,
	)
}
