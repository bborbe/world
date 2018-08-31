package build

import (
	"context"

	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type NginxAutoindex struct {
	Image docker.Image
}

func (t *NginxAutoindex) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Image,
	)
}

func (n *NginxAutoindex) Children() []world.Configuration {
	return []world.Configuration{
		&buildConfiguration{
			&docker.CloneBuilder{
				SourceImage: docker.Image{
					Repository: "jrelva/nginx-autoindex",
					Tag:        "latest",
				},
				TargetImage: n.Image,
			},
		},
	}
}

func (n *NginxAutoindex) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: n.Image,
	}, nil
}
