package build

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
)

type BackupCleanupCron struct {
	Image docker.Image
}

func (t *BackupCleanupCron) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Image,
	)
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

func (b *BackupCleanupCron) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: b.Image,
	}, nil
}
