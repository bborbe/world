// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package apt

import (
	"context"

	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
)

type Update struct {
	SSH *ssh.SSH
}

func (u *Update) Satisfied(ctx context.Context) (bool, error) {
	return false, nil
}

func (u *Update) Apply(ctx context.Context) error {
	return u.SSH.RunCommand(ctx, "DEBIAN_FRONTEND=noninteractive apt-get update --quiet")
}

func (u *Update) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		u.SSH,
	)
}
