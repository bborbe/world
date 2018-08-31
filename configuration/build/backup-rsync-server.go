package build

import (
	"context"

	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type BackupRsyncServer struct {
	Image docker.Image
}

func (t *BackupRsyncServer) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Image,
	)
}

func (b *BackupRsyncServer) Children() []world.Configuration {
	return []world.Configuration{
		&buildConfiguration{
			&docker.Builder{
				GitRepo:   "https://github.com/bborbe/docker-backup-rsync-server.git",
				Image:     b.Image,
				GitBranch: docker.GitBranch(b.Image.Tag),
			},
		},
	}
}

func (b *BackupRsyncServer) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: b.Image,
	}, nil
}
