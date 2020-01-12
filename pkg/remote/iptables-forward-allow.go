// Copyright (c) 2020 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package remote

import (
	"context"

	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/pkg/errors"
)

type IptablesAllowForward struct {
	SSH *ssh.SSH
}

func (i *IptablesAllowForward) Satisfied(ctx context.Context) (bool, error) {
	return i.SSH.RunCommand(ctx, "iptables -C FORWARD -j ACCEPT") == nil, nil
}

func (i *IptablesAllowForward) Apply(ctx context.Context) error {
	return errors.Wrap(i.SSH.RunCommand(ctx, "iptables -A FORWARD -j ACCEPT"), "iptables failed")
}

func (i *IptablesAllowForward) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		i.SSH,
	)
}
