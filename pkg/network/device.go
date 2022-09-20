// Copyright (c) 2021 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package network

import (
	"context"

	"github.com/pkg/errors"
)

type Device string

func (d Device) String() string {
	return string(d)
}
func (d Device) Validate(ctx context.Context) error {
	if d == "" {
		return errors.Errorf("device missing")
	}
	return nil
}
