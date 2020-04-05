// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package build

import (
	"context"

	"github.com/bborbe/world/pkg/build"

	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type PostgresBackup struct {
	Image docker.Image
}

func (p *PostgresBackup) Children() []world.Configuration {
	return []world.Configuration{
		build.Configuration(
			&docker.Builder{
				GitRepo:   "https://github.com/bborbe/postgres-backup.git",
				Image:     p.Image,
				GitBranch: docker.GitBranch(p.Image.Tag),
			},
		),
	}
}

func (p *PostgresBackup) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: p.Image,
	}, nil
}

func (p *PostgresBackup) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		p.Image,
	)
}
