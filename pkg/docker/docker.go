package docker

import (
	"context"
	"os"
	"os/exec"

	"github.com/bborbe/world"
	"github.com/golang/glog"
)

func ImageExists(ctx context.Context, image world.Image) (bool, error) {
	glog.V(1).Infof("check image %s exists locally", image.String())
	cmd := exec.CommandContext(ctx, "docker", "image", "inspect", image.String())
	if glog.V(4) {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		glog.V(1).Infof("image %s not exists locally", image.String())
		return false, nil
	}
	glog.V(1).Infof("image %s exists locally", image.String())
	return true, nil
}
