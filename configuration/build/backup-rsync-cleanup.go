// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package build

import (
	"context"

	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type BackupRsyncCleanup struct {
	Image docker.Image
}

func (t *BackupRsyncCleanup) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Image,
	)
}

func (b *BackupRsyncCleanup) Children() []world.Configuration {
	return []world.Configuration{
		&buildConfiguration{
			&docker.GolangBuilder{
				Name:            "backup-cleanup",
				GitRepo:         "https://github.com/bborbe/backup.git",
				SourceDirectory: "github.com/bborbe/backup",
				Package:         "github.com/bborbe/backup/cmd/backup-cleanup",
				Image:           b.Image,
			},
		},
	}
}

func (b *BackupRsyncCleanup) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: b.Image,
	}, nil
}
