package build

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
)

type PostgresBackup struct {
	Image docker.Image
}

func (b *PostgresBackup) Children() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguration().WithApplier(&docker.Builder{
			GitRepo:   "https://github.com/bborbe/postgres-backup.git",
			Image:     b.Image,
			GitBranch: docker.GitBranch(b.Image.Tag),
		}),
	}
}

func (b *PostgresBackup) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: b.Image,
	}, nil
}

func (t *PostgresBackup) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Image,
	)
}
