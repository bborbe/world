// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service

import (
	"context"

	"github.com/bborbe/world/pkg/apt"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Smartd struct {
	SSH *ssh.SSH
}

func (s *Smartd) Children(ctx context.Context) (world.Configurations, error) {
	return world.Configurations{
		world.NewConfiguraionBuilder().WithApplier(&apt.Install{
			SSH:     s.SSH,
			Package: "smartmontools",
		}),
	}, nil
}

func (s *Smartd) Applier() (world.Applier, error) {
	return nil, nil
}

func (s *Smartd) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		s.SSH,
	)
}
