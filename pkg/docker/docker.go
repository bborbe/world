package docker

import (
	"context"
	"os"
	"os/exec"

	"github.com/bborbe/world"
	"github.com/golang/glog"
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
