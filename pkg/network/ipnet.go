// Copyright (c) 2019 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package network

import (
	"context"
	"net"
)

type IPNet interface {
	IPNet(ctx context.Context) (net.IPNet, error)
	Validate(ctx context.Context) error
}

type IPNetStatic string

func (i IPNetStatic) IPNet(ctx context.Context) (net.IPNet, error) {
	_, ipnet, err := net.ParseCIDR(i.String())
	if err != nil {
		return net.IPNet{}, err
	}
	return *ipnet, nil
}

func (i IPNetStatic) String() string {
	return string(i)
}

func (i IPNetStatic) Validate(ctx context.Context) error {
	return nil
}
