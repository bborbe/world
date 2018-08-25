package build

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
)

type Confluence struct {
	Image         docker.Image
	VendorVersion docker.Tag
	GitBranch     docker.GitBranch
}

func (t *Confluence) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Image,
		t.VendorVersion,
		t.GitBranch,
	)
}

func (c *Confluence) Children() []world.Configuration {
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

func (c *Confluence) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: c.Image,
	}, nil
}
