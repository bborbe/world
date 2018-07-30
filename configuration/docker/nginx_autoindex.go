package docker

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
	"github.com/pkg/errors"
)

type NginxAutoindex struct {
	Image world.Image
}

func (n *NginxAutoindex) Childs() []world.Configuration {
	return []world.Configuration{
		&docker.CloneBuilder{
			SourceImage: world.Image{
				Registry:   "docker.io",
				Repository: "jrelva/nginx-autoindex",
				Tag:        "latest",
			},
			TargetImage: n.Image,
		},
	}
}

func (n *NginxAutoindex) Applier() world.Applier {
	return &docker.Uploader{
		Image: n.Image,
	}
}

func (n *NginxAutoindex) Validate(ctx context.Context) error {
	if err := n.Image.Validate(ctx); err != nil {
		return errors.Wrap(err, "image missing")
	}
	return nil
}
