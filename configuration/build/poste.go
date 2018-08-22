package build

import (
	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
)

type Poste struct {
	Image         docker.Image
	VendorVersion docker.Tag
	GitBranch     docker.GitBranch
}

func (p *Poste) Children() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguration().WithApplier(&docker.Builder{
			GitRepo: "https://github.com/bborbe/poste.io.git",
			Image:   p.Image,
			BuildArgs: docker.BuildArgs{
				"VENDOR_VERSION": p.VendorVersion.String(),
			},
			GitBranch: p.GitBranch,
		}),
	}
}

func (p *Poste) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: p.Image,
	}, nil
}
