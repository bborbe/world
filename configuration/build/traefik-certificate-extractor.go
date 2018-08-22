package build

import (
	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
)

type TraefikCertificateExtractor struct {
	Image docker.Image
}

func (i *TraefikCertificateExtractor) Children() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguration().WithApplier(&docker.Builder{
			GitRepo:   "https://github.com/bborbe/traefik-certificate-extractor.git",
			GitBranch: "master",
			Image:     i.Image,
		}),
	}
}

func (i *TraefikCertificateExtractor) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: i.Image,
	}, nil
}
