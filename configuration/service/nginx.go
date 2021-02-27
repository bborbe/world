// Copyright (c) 2020 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service

import (
	"context"

	"github.com/bborbe/world/pkg/apt"
	"github.com/bborbe/world/pkg/content"
	"github.com/bborbe/world/pkg/file"
	"github.com/bborbe/world/pkg/remote"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Nginx struct {
	SSH *ssh.SSH
}

func (s *Nginx) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		s.SSH,
	)
}

func (s *Nginx) Children() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguraionBuilder().WithApplier(&apt.Install{
			SSH:     s.SSH,
			Package: "nginx",
		}),
		world.NewConfiguraionBuilder().WithApplier(&apt.Install{
			SSH:     s.SSH,
			Package: "certbot",
		}),
		world.NewConfiguraionBuilder().WithApplier(&apt.Install{
			SSH:     s.SSH,
			Package: "python-certbot-nginx",
		}),
		&remote.File{
			SSH:  s.SSH,
			Path: file.Path("/etc/nginx/conf.d/letsencrypt.conf"),
			Content: content.Static(`
`),
			User:  "root",
			Group: "root",
			Perm:  0664,
		},
		world.NewConfiguraionBuilder().WithApplier(&remote.Command{
			SSH:     s.SSH,
			Command: "systemctl restart nginx",
		}),
	}
}

func (s *Nginx) Applier() (world.Applier, error) {
	return nil, nil
}
