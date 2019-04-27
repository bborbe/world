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

type Directory struct {
	SSH   *ssh.SSH
	Path  remote.Path
	User  remote.User
	Group remote.Group
	Perm  remote.Perm
}

func (f *Directory) Children() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguraionBuilder().WithApplier(&remote.Directory{
			SSH:  f.SSH,
			Path: f.Path,
		}),
		world.NewConfiguraionBuilder().WithApplier(&remote.Chown{
			SSH:   f.SSH,
			Path:  f.Path,
			User:  f.User,
			Group: f.Group,
		}),
		world.NewConfiguraionBuilder().WithApplier(&remote.Chmod{
			SSH:  f.SSH,
			Path: f.Path,
			Perm: f.Perm,
		}),
	}
}

func (f *Directory) Applier() (world.Applier, error) {
	return nil, nil
}

func (f *Directory) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		f.SSH,
		f.Path,
		f.User,
		f.Group,
		f.Perm,
	)
}
