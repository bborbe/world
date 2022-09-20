// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service

import (
	"context"

	"github.com/bborbe/world/pkg/remote"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type DisablePostfix struct {
	SSH *ssh.SSH
}

func (d *DisablePostfix) Children(ctx context.Context) (world.Configurations, error) {
	return world.Configurations{}, nil
}

func (d *DisablePostfix) Applier() (world.Applier, error) {
	return &remote.ServiceStop{
		SSH:  d.SSH,
		Name: "postfix",
	}, nil
}

func (d *DisablePostfix) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		d.SSH,
	)
}
