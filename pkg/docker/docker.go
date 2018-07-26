package docker

import (
	"context"

	"os"
	"os/exec"

	"github.com/bborbe/world"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

func ImageExists(ctx context.Context, image world.Image) (bool, error) {
	cmd := exec.CommandContext(ctx, "docker", "image", "inspect", image.String())
	if glog.V(4) {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		return false, nil
	}
	return true, nil
}

func BuildIfNeeded(ctx context.Context, builder world.Builder) error {
	ok, err := builder.Satisfied(ctx)
	if err != nil {
		return errors.Wrap(err, "check satisfied failed")
	}
	if !ok {
		if err := builder.Build(ctx); err != nil {
			return errors.Wrap(err, "build failed")
		}
	}
	return nil
}
