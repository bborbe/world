// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service

import (
	"context"

	"github.com/pkg/errors"

	"github.com/bborbe/world/pkg/file"
	"github.com/bborbe/world/pkg/remote"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Directory struct {
	SSH   *ssh.SSH
	Path  file.HasPath
	User  file.User
	Group file.Group
	Perm  file.Perm
}

func (d *Directory) Children(ctx context.Context) (world.Configurations, error) {
	return world.Configurations{
		world.NewConfiguraionBuilder().WithApplier(&remote.Directory{
			SSH:  d.SSH,
			Path: d.Path,
		}),
		world.NewConfiguraionBuilder().WithApplier(&remote.Chown{
			SSH:   d.SSH,
			Path:  d.Path,
			User:  d.User,
			Group: d.Group,
		}),
		world.NewConfiguraionBuilder().WithApplier(&remote.Chmod{
			SSH:  d.SSH,
			Path: d.Path,
			Perm: d.Perm,
		}),
	}, nil
}

func (d *Directory) Applier() (world.Applier, error) {
	return nil, nil
}

func (d *Directory) Validate(ctx context.Context) error {
	if d.Path == nil {
		return errors.New("Path missing")
	}
	return validation.Validate(
		ctx,
		d.SSH,
		d.User,
		d.Group,
		d.Perm,
	)
}
