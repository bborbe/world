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

func (o ClientName) String() string {
	return string(o)
}
func (o ClientName) Validate(ctx context.Context) error {
	if o == "" {
		return errors.New("ClientName empty")
	}
	return nil
}

type ServerName string

func (o ServerName) String() string {
	return string(o)
}
func (o ServerName) Validate(ctx context.Context) error {
	if o == "" {
		return errors.New("ServerName empty")
	}
	return nil
}

type Routes []Route

func (o Routes) Validate(ctx context.Context) error {
	for _, route := range o {
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

func (o Route) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		o.IPNet,
		o.Gatway,
	)
}
