// Copyright (c) 2019 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package network

import (
	"context"
	"errors"
)

type Host string

func (h Host) String() string {
	return string(h)
}

func (h Host) Validate(ctx context.Context) error {
	if h == "" {
		return errors.New("Host missing")
	}
	return nil
}
