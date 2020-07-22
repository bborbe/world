// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package remote

import (
	"context"

	"github.com/pkg/errors"

	"github.com/bborbe/world/pkg/content"
	"github.com/bborbe/world/pkg/file"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type File struct {
	SSH     *ssh.SSH
	Path    file.HasPath
	Content content.HasContent
	User    file.User
	Group   file.Group
	Perm    file.Perm
}

func (f *File) Children() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguraionBuilder().WithApplier(&FileContent{
			SSH:     f.SSH,
			Path:    f.Path,
			Content: f.Content,
		}),
		world.NewConfiguraionBuilder().WithApplier(&Chown{
			SSH:   f.SSH,
			Path:  f.Path,
			User:  f.User,
			Group: f.Group,
		}),
		world.NewConfiguraionBuilder().WithApplier(&Chmod{
			SSH:  f.SSH,
			Path: f.Path,
			Perm: f.Perm,
		}),
	}
}

func (f *File) Applier() (world.Applier, error) {
	return nil, nil
}

func (f *File) Validate(ctx context.Context) error {
	if f.Path == nil {
		return errors.New("Path missing")
	}
	return validation.Validate(
		ctx,
		f.SSH,
		f.User,
		f.Group,
		f.Perm,
	)
}
