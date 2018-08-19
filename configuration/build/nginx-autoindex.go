package build

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
	"github.com/pkg/errors"
)

type NginxAutoindex struct {
	Image docker.Image
}

func (n *NginxAutoindex) Children() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguration().WithApplier(&docker.CloneBuilder{
			SourceImage: docker.Image{
				Registry:   "docker.io",
				Repository: "jrelva/nginx-autoindex",
				Tag:        "latest",
			},
			TargetImage: n.Image,
		}),
	}
}

func (n *NginxAutoindex) Applier() world.Applier {
	return &docker.Uploader{
		Image: n.Image,
	}
}

func (n *NginxAutoindex) Validate(ctx context.Context) error {
	if err := n.Image.Validate(ctx); err != nil {
		return errors.Wrap(err, "Image missing")
	}
	return nil
}
