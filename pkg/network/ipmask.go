// Copyright (c) 2019 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package network

import (
	"context"
	"net"
	"strconv"
)

type IPMask interface {
	IPMask(ctx context.Context) (net.IPMask, error)
	Validate(ctx context.Context) error
}

type MaskStatic int

func (i MaskStatic) IPMask(ctx context.Context) (net.IPMask, error) {
	return net.CIDRMask(int(i), 32), nil
}

func (i MaskStatic) String() string {
	return strconv.Itoa(i.Int())
}

func (i MaskStatic) Int() int {
	return int(i)
}

func (i MaskStatic) Validate(ctx context.Context) error {
	return nil
}
