package build

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
)

type BackupStatusServer struct {
	Image docker.Image
}

func (t *BackupStatusServer) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Image,
	)
}

func (b *BackupStatusServer) Children() []world.Configuration {
	return []world.Configuration{
		&buildConfiguration{
			&docker.GolangBuilder{
				Name:            "backup-status-server",
				GitRepo:         "https://github.com/bborbe/backup.git",
				SourceDirectory: "github.com/bborbe/backup",
				Package:         "github.com/bborbe/backup/cmd/backup-status-server",
				Image:           b.Image,
			},
		},
	}
}

func (b *BackupStatusServer) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: b.Image,
	}, nil
}
