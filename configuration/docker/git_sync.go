package docker

import (
	"context"

	"github.com/bborbe/world"
	"github.com/pkg/errors"
)

type GitSync struct {
	Image world.Image
}

func (n *GitSync) Childs() []world.Configuration {
	return nil
}

func (n *GitSync) Applier() world.Applier {
	return nil
}

func (n *GitSync) Validate(ctx context.Context) error {
	if err := n.Image.Validate(ctx); err != nil {
		return errors.Wrap(err, "image missing")
	}
	return nil
}
