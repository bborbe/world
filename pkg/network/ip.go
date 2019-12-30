// Copyright (c) 2019 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package network

import (
	"context"
	"fmt"
	"net"
)

type IP interface {
	IP(ctx context.Context) (net.IP, error)
	Validate(ctx context.Context) error
}

type IPStatic string

func (i IPStatic) IP(ctx context.Context) (net.IP, error) {
	return net.ParseIP(i.String()), nil
}

func (i IPStatic) String() string {
	return string(i)
}

func (i IPStatic) Validate(ctx context.Context) error {
	ip, err := i.IP(ctx)
	if err != nil {
		return err
	}
	if !ip.To4().Equal(ip) {
		return fmt.Errorf("ip '%s' is not a ipv4 addr", i.String())
	}
	return nil
}
