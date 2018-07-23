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
	Builder world.Builder
}

func (b *Uploader) Upload(ctx context.Context) error {
	glog.V(2).Infof("run docker build ...")
	image := b.GetBuilder().GetImage()
	cmd := exec.CommandContext(ctx, "docker", "push", image.String())
	if glog.V(4) {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return errors.Wrap(cmd.Run(), "upload docker image failed")
}

func (b *Uploader) Validate(ctx context.Context) error {
	if b.Builder == nil {
		return errors.New("build missing")
	}
	if err := b.Builder.Validate(ctx); err != nil {
		return err
	}
	return nil
}

func (b *Uploader) GetBuilder() world.Builder {
	return b.Builder
}
