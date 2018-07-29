package docker

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
	"github.com/pkg/errors"
)

type Download struct {
	Image world.Image
}

func (d *Download) Childs() []world.Configuration {
	return []world.Configuration{
		&docker.CloneBuilder{
			SourceImage: world.Image{
				Registry:   "docker.io",
				Repository: "jrelva/nginx-autoindex",
				Tag:        "latest",
			},
			TargetImage: d.Image,
		},
	}
}

func (d *Download) Applier() world.Applier {
	return &docker.Uploader{
		Image: d.Image,
	}
}

func (d *Download) Validate(ctx context.Context) error {
	if err := d.Image.Validate(ctx); err != nil {
		return errors.Wrap(err, "image missing")
	}
	return nil
}
