package build

import (
	"context"

	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type HostPathProvisioner struct {
	Image docker.Image
}

func (t *HostPathProvisioner) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Image,
	)
}

func (i *HostPathProvisioner) Children() []world.Configuration {
	return []world.Configuration{
		&buildConfiguration{
			&docker.GolangBuilder{
				Name:            "hostpath-provisioner",
				GitRepo:         "https://github.com/bborbe/hostpath-provisioner.git",
				SourceDirectory: "github.com/bborbe/hostpath-provisioner",
				Package:         "github.com/bborbe/hostpath-provisioner",
				Image:           i.Image,
			},
		},
	}
}

func (i *HostPathProvisioner) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: i.Image,
	}, nil
}
