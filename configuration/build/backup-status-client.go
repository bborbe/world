package build

import (
	"context"

	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type BackupStatusClient struct {
	Image docker.Image
}

func (t *BackupStatusClient) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Image,
	)
}

func (b *BackupStatusClient) Children() []world.Configuration {
	return []world.Configuration{
		&buildConfiguration{
			&docker.GolangBuilder{
				Name:            "backup-status-client",
				GitRepo:         "https://github.com/bborbe/backup.git",
				SourceDirectory: "github.com/bborbe/backup",
				Package:         "github.com/bborbe/backup/cmd/backup-status-client",
				Image:           b.Image,
			},
		},
	}
}

func (b *BackupStatusClient) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: b.Image,
	}, nil
}