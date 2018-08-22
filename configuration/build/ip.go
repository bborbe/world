package build

import (
	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
)

type Ip struct {
	Image docker.Image
}

func (i *Ip) Children() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguration().WithApplier(&docker.GolangBuilder{
			Name:            "ip",
			GitRepo:         "https://github.com/bborbe/ip.git",
			SourceDirectory: "github.com/bborbe/ip",
			Package:         "github.com/bborbe/ip/cmd/ip-server",
			Image:           i.Image,
		}),
	}
}

func (i *Ip) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: i.Image,
	}, nil
}
