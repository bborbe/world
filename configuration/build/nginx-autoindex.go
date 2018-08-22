package build

import (
	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
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

func (n *NginxAutoindex) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: n.Image,
	}, nil
}
