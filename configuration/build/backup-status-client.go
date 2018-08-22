package build

import (
	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
)

type BackupStatusClient struct {
	Image docker.Image
}

func (b *BackupStatusClient) Children() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguration().WithApplier(&docker.GolangBuilder{
			Name:            "backup-status-client",
			GitRepo:         "https://github.com/bborbe/backup.git",
			SourceDirectory: "github.com/bborbe/backup",
			Package:         "github.com/bborbe/backup/cmd/backup-status-client",
			Image:           b.Image,
		}),
	}
}

func (b *BackupStatusClient) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: b.Image,
	}, nil
}
