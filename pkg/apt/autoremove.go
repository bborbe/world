// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package apt

import (
	"context"

	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
)

type Autoremove struct {
	SSH *ssh.SSH
}

func (a *Autoremove) Satisfied(ctx context.Context) (bool, error) {
	return false, nil
}

func (a *Autoremove) Apply(ctx context.Context) error {
	return a.SSH.RunCommand(ctx, "DEBIAN_FRONTEND=noninteractive apt-get autoremove --yes")
}

func (a *Autoremove) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		a.SSH,
	)
}
