// Copyright (c) 2019 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package openvpn

import (
	"context"

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

type Routes []Route

func (r Routes) Validate(ctx context.Context) error {
	for _, route := range r {
		if err := route.Validate(ctx); err != nil {
			return err
		}
	}
	return nil
}

type Route struct {
	Gatway network.IP
	IPNet  network.IPNet
}

func (r Route) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		r.IPNet,
		r.Gatway,
	)
}
