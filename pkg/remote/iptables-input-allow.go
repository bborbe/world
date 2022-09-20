// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package remote

import (
	"context"
	"fmt"
	"strconv"

	"github.com/pkg/errors"

	"github.com/bborbe/world/pkg/network"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
)

type IptablesAllowInput struct {
	SSH       *ssh.SSH
	Port      network.Port
	PortRange *network.PortRange
	Protocol  network.Protocol
}

func (i *IptablesAllowInput) Satisfied(ctx context.Context) (bool, error) {
	portString, err := i.portString(ctx)
	if err != nil {
		return false, err
	}
	return i.SSH.RunCommand(ctx, fmt.Sprintf("iptables -C INPUT -p %s -m state --state NEW -m %s --dport %s -j ACCEPT", i.Protocol, i.Protocol, portString)) == nil, nil
}

func (i *IptablesAllowInput) Apply(ctx context.Context) error {
	portString, err := i.portString(ctx)
	if err != nil {
		return err
	}
	return errors.Wrap(i.SSH.RunCommand(ctx, fmt.Sprintf("iptables -A INPUT -p %s -m state --state NEW -m %s --dport %s -j ACCEPT", i.Protocol, i.Protocol, portString)), "iptables failed")
}

func (i *IptablesAllowInput) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		i.SSH,
		i.Protocol,
		// TODO: WTF?
		//validation.EitherValidation(
		//	i.PortRange,
		//	i.Port,
		//),
	)
}

func (i *IptablesAllowInput) portString(ctx context.Context) (string, error) {
	if i.Port != nil {
		port, err := i.Port.Port(ctx)
		if err != nil {
			return "", errors.Wrap(err, "get port failed")
		}
		return strconv.Itoa(port), nil
	}
	if i.PortRange != nil {
		from, err := i.PortRange.From.Port(ctx)
		if err != nil {
			return "", errors.Wrap(err, "get port failed")
		}
		to, err := i.PortRange.To.Port(ctx)
		if err != nil {
			return "", errors.Wrap(err, "get port failed")
		}
		return fmt.Sprintf("%d:%d", from, to), nil
	}
	return "", errors.Errorf("port and portrange empty")
}
