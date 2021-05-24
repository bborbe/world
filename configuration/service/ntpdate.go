// Copyright (c) 2020 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service

import (
	"context"
	"github.com/bborbe/world/pkg/content"

	"github.com/bborbe/world/pkg/apt"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type NtpDate struct {
	SSH *ssh.SSH
}

func (d *NtpDate) Children() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguraionBuilder().WithApplier(&apt.Install{
			SSH:     d.SSH,
			Package: "ntpdate",
		}),
		&Cron{
			SSH:        d.SSH,
			Name:       "ntpdate",
			Expression: content.Static("ntpdate -s de.pool.ntp.org > /dev/null"),
			Schedule:   "15 * * * *",
		},
	}
}

func (d *NtpDate) Applier() (world.Applier, error) {
	return nil, nil
}

func (d *NtpDate) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		d.SSH,
	)
}
