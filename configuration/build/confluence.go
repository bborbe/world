package build

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
	"github.com/pkg/errors"
)

type Confluence struct {
	Image         docker.Image
	VendorVersion docker.Tag
	GitBranch     docker.GitBranch
}

func (c *Confluence) Childs() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguration().WithApplier(&docker.Builder{
			GitRepo: "https://github.com/bborbe/atlassian-confluence.git",
			Image:   c.Image,
			BuildArgs: docker.BuildArgs{
				"VENDOR_VERSION": c.VendorVersion.String(),
			},
			GitBranch: c.GitBranch,
		}),
	}
}

func (c *Confluence) Applier() world.Applier {
	return &docker.Uploader{
		Image: c.Image,
	}
}

func (c *Confluence) Validate(ctx context.Context) error {
	if err := c.Image.Validate(ctx); err != nil {
		return errors.Wrap(err, "image missing")
	}
	return nil
}
