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

type BackupStatusServer struct {
	Image docker.Image
}

func (b *BackupStatusServer) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		b.Image,
	)
}

func (b *BackupStatusServer) Children(ctx context.Context) (world.Configurations, error) {
	return world.Configurations{
		build.Configuration(
			&docker.GolangBuilder{
				Name:            "backup-status-server",
				GitRepo:         "https://github.com/bborbe/backup.git",
				SourceDirectory: "github.com/bborbe/backup",
				Package:         "github.com/bborbe/backup/cmd/backup-status-server",
				Image:           b.Image,
			},
		),
	}, nil
}

func (b *BackupStatusServer) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: b.Image,
	}, nil
}
