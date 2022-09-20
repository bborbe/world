// Copyright (c) 2019 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package network

import (
	"context"

	"github.com/bborbe/world/pkg/validation"
)

type PortRange struct {
	From Port
	To   Port
}

func (p PortRange) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		p.To,
		p.From,
	)
}
