// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package remote

import (
	"context"
	"fmt"

	"github.com/bborbe/world/pkg/file"

	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/pkg/errors"
)

type Directory struct {
	SSH  *ssh.SSH
	Path file.HasPath
}

func (d *Directory) Satisfied(ctx context.Context) (bool, error) {
	path, err := d.Path.Path(ctx)
	if err != nil {
		return false, err
	}
	return d.SSH.RunCommand(ctx, fmt.Sprintf("test -d %s", path)) == nil, nil
}

func (d *Directory) Apply(ctx context.Context) error {
	path, err := d.Path.Path(ctx)
	if err != nil {
		return err
	}
	return errors.Wrap(d.SSH.RunCommand(ctx, fmt.Sprintf("mkdir -p %s", path)), "mkdir failed")
}

func (d *Directory) Validate(ctx context.Context) error {
	if d.Path == nil {
		return errors.New("Path missing")
	}
	return validation.Validate(
		ctx,
		d.SSH,
	)
}
