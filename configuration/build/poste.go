package build

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
	"github.com/pkg/errors"
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

func (p *Poste) Applier() world.Applier {
	return &docker.Uploader{
		Image: p.Image,
	}
}

func (p *Poste) Validate(ctx context.Context) error {
	if err := p.Image.Validate(ctx); err != nil {
		return errors.Wrap(err, "Image missing")
	}
	if p.GitBranch == "" {
		return errors.New("GitBranch missing")
	}
	if p.VendorVersion == "" {
		return errors.New("VendorVersion missing")
	}
	return nil
}
