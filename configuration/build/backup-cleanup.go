package build

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
	"github.com/pkg/errors"
)

type BackupCleanupCron struct {
	Image docker.Image
}

func (b *BackupCleanupCron) Children() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguration().WithApplier(&docker.Builder{
			GitRepo:   "https://github.com/bborbe/backup-cleanup-cron.git",
			Image:     b.Image,
			GitBranch: docker.GitBranch(b.Image.Tag),
		}),
	}
}

func (b *BackupCleanupCron) Applier() world.Applier {
	return &docker.Uploader{
		Image: b.Image,
	}
}

func (b *BackupCleanupCron) Validate(ctx context.Context) error {
	if err := b.Image.Validate(ctx); err != nil {
		return errors.Wrap(err, "Image missing")
	}
	return nil
}
