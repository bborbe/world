package build

import (
	"context"

	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type TraefikCertificateExtractor struct {
	Image docker.Image
}

func (w *TraefikCertificateExtractor) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		w.Image,
	)
}

func (i *TraefikCertificateExtractor) Children() []world.Configuration {
	return []world.Configuration{
		&buildConfiguration{
			&docker.Builder{
				GitRepo:   "https://github.com/bborbe/traefik-certificate-extractor.git",
				GitBranch: "master",
				Image:     i.Image,
			},
		},
	}
}

func (i *TraefikCertificateExtractor) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: i.Image,
	}, nil
}
