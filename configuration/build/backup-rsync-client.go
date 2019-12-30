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

type BackupRsyncClient struct {
	Image docker.Image
}

func (b *BackupRsyncClient) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		b.Image,
	)
}

func (b *BackupRsyncClient) Children() []world.Configuration {
	return []world.Configuration{
		&buildConfiguration{
			&docker.Builder{
				GitRepo: "https://github.com/bborbe/backup-rsync-client.git",
				BuildArgs: docker.BuildArgs{
					"VENDOR_VERSION": "2.0.0",
				},
				Image:     b.Image,
				GitBranch: docker.GitBranch(b.Image.Tag),
			},
		},
	}
}

func (b *BackupRsyncClient) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: b.Image,
	}, nil
}
