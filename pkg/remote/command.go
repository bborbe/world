// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package remote

import (
	"context"

	"github.com/pkg/errors"

	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
)

type Command struct {
	SSH *ssh.SSH

	Command string
}

func (f *Command) Satisfied(ctx context.Context) (bool, error) {
	return false, nil
}

func (f *Command) Apply(ctx context.Context) error {
	return f.SSH.RunCommand(ctx, f.Command)
}

func (f *Command) Validate(ctx context.Context) error {
	if f.Command == "" {
		return errors.New("Command missing")
	}
	return validation.Validate(
		ctx,
		f.SSH,
	)
}
