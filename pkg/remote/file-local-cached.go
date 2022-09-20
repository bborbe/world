// Copyright (c) 2019 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package remote

import (
	"context"
	"io/ioutil"

	"github.com/pkg/errors"

	"github.com/bborbe/world/pkg/content"
	"github.com/bborbe/world/pkg/file"
	"github.com/bborbe/world/pkg/local"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type FileLocalCached struct {
	SSH       *ssh.SSH
	Path      file.HasPath
	LocalPath file.HasPath
	Content   content.HasContent
	User      file.User
	Group     file.Group
	Perm      file.Perm
}

func (f *FileLocalCached) Children(ctx context.Context) (world.Configurations, error) {
	return world.Configurations{
		world.NewConfiguraionBuilder().WithApplier(
			&local.FileContent{
				Path:    f.LocalPath,
				Content: f.Content,
			},
		),
		&File{
			SSH:  f.SSH,
			Path: f.Path,
			Content: content.Func(func(ctx context.Context) ([]byte, error) {
				path, err := f.LocalPath.Path(ctx)
				if err != nil {
					return nil, err
				}
				return ioutil.ReadFile(path)
			}),
			User:  f.User,
			Group: f.Group,
			Perm:  f.Perm,
		},
	}, nil
}

func (f *FileLocalCached) Applier() (world.Applier, error) {
	return nil, nil
}

func (f *FileLocalCached) Validate(ctx context.Context) error {
	if f.Path == nil {
		return errors.New("Path missing")
	}
	if f.LocalPath == nil {
		return errors.New("LocalPath missing")
	}
	return validation.Validate(
		ctx,
		f.SSH,
		f.User,
		f.Group,
		f.Perm,
	)
}
