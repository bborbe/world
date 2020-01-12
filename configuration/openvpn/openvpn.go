// Copyright (c) 2019 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package openvpn

import (
	"context"

	"github.com/bborbe/world/configuration/server"
	"github.com/bborbe/world/pkg/network"
	"github.com/bborbe/world/pkg/validation"
	"github.com/pkg/errors"
)

type ClientName string

func (c ClientName) String() string {
	return string(c)
}
func (c ClientName) Validate(ctx context.Context) error {
	if c == "" {
		return errors.New("ClientName empty")
	}
	return nil
}

type ServerName string

func (s ServerName) String() string {
	return string(s)
}
func (s ServerName) Validate(ctx context.Context) error {
	if s == "" {
		return errors.New("ServerName empty")
	}
	return nil
}

type ServerAddress string

func (s ServerAddress) String() string {
	return string(s)
}
func (s ServerAddress) Validate(ctx context.Context) error {
	if s == "" {
		return errors.New("ServerName empty")
	}
	return nil
}

type ServerRoutes []ServerRoute

func (r ServerRoutes) Validate(ctx context.Context) error {
	for _, route := range r {
		if err := route.Validate(ctx); err != nil {
			return err
		}
	}
	return nil
}

type ServerRoute struct {
	Gatway network.IP
	IPNet  network.IPNet
}

func (r ServerRoute) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		r.IPNet,
		r.Gatway,
	)
}

type ClientRoutes []ClientRoute

func (r ClientRoutes) Validate(ctx context.Context) error {
	for _, route := range r {
		if err := route.Validate(ctx); err != nil {
			return err
		}
	}
	return nil
}

type ClientRoute struct {
	IPNet network.IPNet
}

func (r ClientRoute) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		r.IPNet,
	)
}

func BuildClientIPs(servers ...server.Server) ClientIPs {
	var result ClientIPs
	for _, server := range servers {
		result = append(result, ClientIP{
			Name: ClientName(server.Name),
			IP:   server.VpnIP,
		})
	}

	return result
}

type ClientIPs []ClientIP

func (c ClientIPs) Validate(ctx context.Context) error {
	for _, clientIP := range c {
		if err := clientIP.Validate(ctx); err != nil {
			return err
		}
	}
	return nil
}

type ClientIP struct {
	Name ClientName
	IP   network.IP
}

func (c ClientIP) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		c.Name,
		c.IP,
	)
}

func BuildServerRoutes(servers ...server.Server) ServerRoutes {
	var result ServerRoutes
	for _, server := range servers {
		result = append(result, ServerRoute{
			Gatway: server.VpnIP,
			IPNet: network.IPNetFromIP{
				IP:   server.IP,
				Mask: 32,
			},
		})
	}
	return result
}
