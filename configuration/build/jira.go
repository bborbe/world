package build

import (
	"context"

	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Jira struct {
	Image         docker.Image
	VendorVersion docker.Tag
	GitBranch     docker.GitBranch
}

func (t *Jira) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Image,
		t.VendorVersion,
		t.GitBranch,
	)
}

func (c *Jira) Children() []world.Configuration {
	return []world.Configuration{
		&buildConfiguration{
			&docker.Builder{
				GitRepo: "https://github.com/bborbe/atlassian-jira-software.git",
				Image:   c.Image,
				BuildArgs: docker.BuildArgs{
					"VENDOR_VERSION": c.VendorVersion.String(),
				},
				GitBranch: c.GitBranch,
			},
		},
	}
}

func (c *Jira) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: c.Image,
	}, nil
}
