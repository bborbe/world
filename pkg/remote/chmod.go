// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package remote

import (
	"context"
	"fmt"
	"strings"

	"github.com/bborbe/world/pkg/file"

	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/pkg/errors"
)

type Chmod struct {
	SSH *ssh.SSH

	Path file.HasPath
	Perm file.Perm
}

func (c *Chmod) Satisfied(ctx context.Context) (bool, error) {
	path, err := c.Path.Path(ctx)
	if err != nil {
		return false, err
	}
	stdout, err := c.SSH.RunCommandStdout(ctx, "stat -c '%a' "+path)
	if err != nil {
		return false, errors.Wrapf(err, "check stat of %s failed", path)
	}
	return strings.TrimSpace(string(stdout)) == fmt.Sprintf("%o", c.Perm), nil
}

func (c *Chmod) Apply(ctx context.Context) error {
	path, err := c.Path.Path(ctx)
	if err != nil {
		return err
	}
	return errors.Wrap(c.SSH.RunCommand(ctx, fmt.Sprintf("chmod %s %s", c.Perm, path)), "chown failed")
}

func (c *Chmod) Validate(ctx context.Context) error {
	if c.Path == nil {
		return errors.New("Path missing")
	}
	return validation.Validate(
		ctx,
		c.SSH,
		c.Perm,
	)
}
