package build

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
	"github.com/pkg/errors"
)

type BackupRsyncServer struct {
	Image docker.Image
}

func (b *BackupRsyncServer) Children() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguration().WithApplier(&docker.Builder{
			GitRepo:   "https://github.com/bborbe/docker-backup-rsync-server.git",
			Image:     b.Image,
			GitBranch: docker.GitBranch(b.Image.Tag),
		}),
	}
}

func (b *BackupRsyncServer) Applier() world.Applier {
	return &docker.Uploader{
		Image: b.Image,
	}
}

func (b *BackupRsyncServer) Validate(ctx context.Context) error {
	if err := b.Image.Validate(ctx); err != nil {
		return errors.Wrap(err, "Image missing")
	}
	return nil
}
