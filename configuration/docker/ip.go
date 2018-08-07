package docker

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/configuration"
	"github.com/bborbe/world/pkg/docker"
	"github.com/pkg/errors"
)

type Ip struct {
	Image world.Image
}

func (i *Ip) Childs() []world.Configuration {
	return []world.Configuration{
		configuration.New().WithApplier(&docker.GolangBuilder{
			Name:            "ip",
			GitRepo:         "https://github.com/bborbe/ip.git",
			SourceDirectory: "github.com/bborbe/ip",
			Package:         "github.com/bborbe/ip/cmd/ip-server",
			Image:           i.Image,
		}),
	}
}

func (i *Ip) Applier() world.Applier {
	return &docker.Uploader{
		Image: i.Image,
	}
}

func (i *Ip) Validate(ctx context.Context) error {
	if err := i.Image.Validate(ctx); err != nil {
		return errors.Wrap(err, "image missing")
	}
	return nil
}
