package uploader

import (
	"context"
	"os"
	"os/exec"

	"github.com/bborbe/world"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type Uploader struct {
	Image world.Image
}

func (b *Uploader) Upload(ctx context.Context) error {
	glog.V(2).Infof("run docker build ...")
	cmd := exec.CommandContext(ctx, "docker", "push", b.Image.String())
	if glog.V(4) {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return errors.Wrap(cmd.Run(), "upload docker image failed")
}
