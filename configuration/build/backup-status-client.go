package build

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
	"github.com/pkg/errors"
)

type BackupStatusClient struct {
	VendorVersion docker.Tag
	Image         docker.Image
	GitBranch     docker.GitBranch
}

func (b *BackupStatusClient) Children() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguration().WithApplier(&docker.Builder{
			GitRepo: "https://github.com/bborbe/docker-backup-rsync-client.git",
			Image:   b.Image,
			BuildArgs: docker.BuildArgs{
				"VENDOR_VERSION": b.VendorVersion.String(),
			},
			GitBranch: b.GitBranch,
		}),
	}
}

func (b *BackupStatusClient) Applier() world.Applier {
	return &docker.Uploader{
		Image: b.Image,
	}
}

func (b *BackupStatusClient) Validate(ctx context.Context) error {
	if err := b.Image.Validate(ctx); err != nil {
		return errors.Wrap(err, "Image missing")
	}
	if b.GitBranch == "" {
		return errors.New("GitBranch missing")
	}
	if b.VendorVersion == "" {
		return errors.New("VendorVersion missing")
	}
	return nil
}
