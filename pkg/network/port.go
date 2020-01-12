// Copyright (c) 2019 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package network

import (
	"context"

	"github.com/pkg/errors"
)

type Port interface {
	Port(ctx context.Context) (int, error)
	Validate(ctx context.Context) error
}

type PortStatic int

func (i PortStatic) Port(ctx context.Context) (int, error) {
	return i.Int(), nil
}

func (i PortStatic) Int() int {
	return int(i)
}

func (i PortStatic) Validate(ctx context.Context) error {
	if i.Int() <= 0 || i.Int() > 65535 {
		return errors.Errorf("invalid port %d", i.Int())
	}
	return nil
}
