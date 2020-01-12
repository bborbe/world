// Copyright (c) 2019 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package network

import (
	"context"
	"net"
	"strconv"
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

type IPNetFromIP struct {
	IP   IP
	Mask int
}

func (i IPNetFromIP) IPNet(ctx context.Context) (net.IPNet, error) {
	ip, err := i.IP.IP(ctx)
	if err != nil {
		return net.IPNet{}, err
	}
	_, ipnet, err := net.ParseCIDR(ip.String() + "/" + strconv.Itoa(i.Mask))
	if err != nil {
		return net.IPNet{}, err
	}
	return *ipnet, nil

}
func (i IPNetFromIP) Validate(ctx context.Context) error {
	return i.IP.Validate(ctx)
}
