package build

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
	"github.com/pkg/errors"
)

type Jira struct {
	Image         docker.Image
	VendorVersion docker.Tag
	GitBranch     docker.GitBranch
}

func (c *Jira) Children() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguration().WithApplier(&docker.Builder{
			GitRepo: "https://github.com/bborbe/atlassian-jira-software.git",
			Image:   c.Image,
			BuildArgs: docker.BuildArgs{
				"VENDOR_VERSION": c.VendorVersion.String(),
			},
			GitBranch: c.GitBranch,
		}),
	}
}

func (c *Jira) Applier() world.Applier {
	return &docker.Uploader{
		Image: c.Image,
	}
}

func (c *Jira) Validate(ctx context.Context) error {
	if err := c.Image.Validate(ctx); err != nil {
		return errors.Wrap(err, "Image missing")
	}
	if c.GitBranch == "" {
		return errors.New("GitBranch missing")
	}
	if c.VendorVersion == "" {
		return errors.New("VendorVersion missing")
	}
	return nil
}
