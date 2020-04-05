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

type BackupRsyncServer struct {
	Image docker.Image
}

func (b *BackupRsyncServer) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		b.Image,
	)
}

func (b *BackupRsyncServer) Children() []world.Configuration {
	return []world.Configuration{
		build.Configuration(
			&docker.Builder{
				GitRepo:   "https://github.com/bborbe/backup-rsync-server.git",
				Image:     b.Image,
				GitBranch: docker.GitBranch(b.Image.Tag),
			},
		),
	}
}

func (b *BackupRsyncServer) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: b.Image,
	}, nil
}
