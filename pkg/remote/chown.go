// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package remote

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/bborbe/world/pkg/file"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
)

type Chown struct {
	SSH *ssh.SSH

	Path  file.HasPath
	User  file.User
	Group file.Group
}

func (c *Chown) Satisfied(ctx context.Context) (bool, error) {
	path, err := c.Path.Path(ctx)
	if err != nil {
		return false, err
	}
	stdout, err := c.SSH.RunCommandStdout(ctx, "stat -c '%U:%G' "+path)
	if err != nil {
		return false, errors.Wrapf(err, "check stat of %s failed", path)
	}
	return strings.TrimSpace(string(stdout)) == fmt.Sprintf("%s:%s", c.User, c.Group), nil
}

func (c *Chown) Apply(ctx context.Context) error {
	path, err := c.Path.Path(ctx)
	if err != nil {
		return err
	}
	return errors.Wrap(c.SSH.RunCommand(ctx, fmt.Sprintf("chown %s:%s %s", c.User, c.Group, path)), "chown failed")
}

func (c *Chown) Validate(ctx context.Context) error {
	if c.Path == nil {
		return errors.New("Path missing")
	}
	return validation.Validate(
		ctx,
		c.SSH,
		c.User,
		c.Group,
	)
}
