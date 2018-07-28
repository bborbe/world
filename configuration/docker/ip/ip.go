package ip

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
	return nil
}

func (b *Builder) Validate(ctx context.Context) error {
	if err := b.Image.Validate(ctx); err != nil {
		return errors.Wrap(err, "image missing")
	}
	return nil
}

func (b *Builder) Satisfied(ctx context.Context) (bool, error) {
	return false, nil
}

func (b *Builder) Apply(ctx context.Context) error {
	builder := &docker.GolangBuilder{
		Name:            "ip",
		GitRepo:         "https://github.com/bborbe/ip.git",
		SourceDirectory: "github.com/bborbe/ip",
		Package:         "github.com/bborbe/ip/cmd/ip-server",
		Image:           b.Image,
	}
	return builder.Apply(ctx)
}
