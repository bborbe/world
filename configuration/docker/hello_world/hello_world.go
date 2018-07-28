package hello_world

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
	"github.com/pkg/errors"
)

type Builder struct {
	Image world.Image
}

func (b *Builder) Required() world.Applier {
	return &docker.Builder{
		GitRepo: "https://github.com/bborbe/hello-world.git",
		Image:   b.Image,
	}
}

func (b *Builder) Validate(ctx context.Context) error {
	if err := b.Image.Validate(ctx); err != nil {
		return errors.Wrap(err, "image missing")
	}
	return nil
}

func (b *Builder) uploader() *docker.Uploader {
	return &docker.Uploader{
		Image: b.Image,
	}
}
func (b *Builder) Satisfied(ctx context.Context) (bool, error) {
	return b.uploader().Satisfied(ctx)
}

func (b *Builder) Apply(ctx context.Context) error {
	return b.uploader().Apply(ctx)
}
