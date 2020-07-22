// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package local

import (
	"context"
	"os/exec"

	"github.com/pkg/errors"

	"github.com/bborbe/world/pkg/validation"
)

type Command struct {
	Command string
	Args    []string
}

func (c *Command) Satisfied(ctx context.Context) (bool, error) {
	return false, nil
}

func (c *Command) Apply(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, c.Command, c.Args...)
	return errors.Wrapf(cmd.Run(), "execute command %s %v failed", c.Command, c.Args)
}

func (c *Command) Validate(ctx context.Context) error {
	if c.Command == "" {
		return errors.New("Command missing")
	}
	return validation.Validate(
		ctx,
	)
}
