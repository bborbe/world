// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package remote

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/bborbe/world/pkg/network"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
)

type IptablesAllowInput struct {
	SSH  *ssh.SSH
	Port network.Port
}

func (i *IptablesAllowInput) Satisfied(ctx context.Context) (bool, error) {
	port, err := i.Port.Port(ctx)
	if err != nil {
		return false, err
	}
	return i.SSH.RunCommand(ctx, fmt.Sprintf("iptables -C INPUT -p tcp -m state --state NEW -m tcp --dport %d -j ACCEPT", port)) == nil, nil
}

func (i *IptablesAllowInput) Apply(ctx context.Context) error {
	port, err := i.Port.Port(ctx)
	if err != nil {
		return err
	}
	return errors.Wrap(i.SSH.RunCommand(ctx, fmt.Sprintf("iptables -A INPUT -p tcp -m state --state NEW -m tcp --dport %d -j ACCEPT", port)), "iptables failed")
}

func (i *IptablesAllowInput) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		i.SSH,
		i.Port,
	)
}
