// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package build

import (
	"context"

	"github.com/bborbe/world/pkg/world"
)

type buildConfiguration struct {
	applier world.Applier
}

func (c *buildConfiguration) Applier() (world.Applier, error) {
	return c.applier, nil
}

func (c *buildConfiguration) Children() []world.Configuration {
	return nil
}

func (c *buildConfiguration) Validate(ctx context.Context) error {
	return nil
}
