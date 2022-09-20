// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service

import (
	"context"

	"github.com/bborbe/world/pkg/content"
	"github.com/bborbe/world/pkg/file"
	"github.com/bborbe/world/pkg/remote"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type HdParam struct {
	SSH *ssh.SSH
}

func (h *HdParam) Children(ctx context.Context) (world.Configurations, error) {
	return world.Configurations{
		&remote.File{
			SSH:  h.SSH,
			Path: file.Path("/etc/udev/rules.d/69-hdparm.rules\n"),
			Content: content.Static(`ACTION=="add|change", KERNEL=="sd[a-z]", ATTRS{queue/rotational}=="1", RUN+="/usr/sbin/hdparm -S 60 /dev/%k"
`),
			User:  "root",
			Group: "root",
			Perm:  0644,
		},
		world.NewConfiguraionBuilder().WithApplier(&remote.Command{
			SSH:     h.SSH,
			Command: `udevadm control --reload-rules && udevadm trigger`,
		}),
	}, nil
}

func (h *HdParam) Applier() (world.Applier, error) {
	return nil, nil
}

func (h *HdParam) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		h.SSH,
	)
}
