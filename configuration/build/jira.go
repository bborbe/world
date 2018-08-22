package build

import (
	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
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

func (c *Jira) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: c.Image,
	}, nil
}
