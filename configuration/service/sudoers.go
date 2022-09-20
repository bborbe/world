// Copyright (c) 2021 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service

import (
	"context"

	"github.com/bborbe/world/pkg/content"
	"github.com/bborbe/world/pkg/file"
	"github.com/bborbe/world/pkg/remote"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Sudoers struct {
	SSH *ssh.SSH
}

func (d *Sudoers) Children(ctx context.Context) (world.Configurations, error) {
	return world.Configurations{
		&remote.File{
			SSH:     d.SSH,
			Path:    file.Path("/etc/sudoers.d/users"),
			Content: content.Static("bborbe ALL=(ALL) NOPASSWD: ALL"),
			User:    "root",
			Group:   "root",
			Perm:    0440,
		},
	}, nil
}

func (d *Sudoers) Applier() (world.Applier, error) {
	return nil, nil
}

func (d *Sudoers) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		d.SSH,
	)
}
