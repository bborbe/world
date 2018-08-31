package build

import (
	"context"

	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type KubednsMasq struct {
	Image docker.Image
}

func (t *KubednsMasq) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Image,
	)
}

func (n *KubednsMasq) Children() []world.Configuration {
	return []world.Configuration{
		&buildConfiguration{
			&docker.CloneBuilder{
				SourceImage: docker.Image{
					Repository: "gcr.io/google_containers/kube-dnsmasq-amd64",
					Tag:        n.Image.Tag,
				},
				TargetImage: n.Image,
			},
		},
	}
}

func (n *KubednsMasq) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: n.Image,
	}, nil
}
