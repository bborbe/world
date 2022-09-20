// Copyright (c) 2021 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package network

import (
	"context"

	"github.com/pkg/errors"
)

type Protocol string

func (p Protocol) String() string {
	return string(p)
}

func (p Protocol) Validate(ctx context.Context) error {
	switch p {
	case UDP, TCP:
		return nil
	default:
		return errors.Errorf("invalid protocol '%s'", p)
	}
}

const (
	UDP Protocol = "udp"
	TCP Protocol = "tcp"
)
