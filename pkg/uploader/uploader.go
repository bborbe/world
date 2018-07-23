package uploader

import (
	"context"
	"os"
	"os/exec"

	"fmt"
	"net/http"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/builder"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

func UploadIfNeeded(ctx context.Context, uploader world.Uploader) error {
	ok, err := uploader.Satisfied(ctx)
	if err != nil {
		return errors.Wrap(err, "check satisfied failed")
	}
	if !ok {
		if err := uploader.Upload(ctx); err != nil {
			return errors.Wrap(err, "upload failed")
		}
	}
	return nil
}

type Uploader struct {
	Builder world.Builder
}

func (u *Uploader) Upload(ctx context.Context) error {
	image := u.GetBuilder().GetImage()
	glog.V(2).Infof("upload docker image %s ...", image.String())
	if err := builder.BuildIfNeeded(ctx, u.Builder); err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, "docker", "push", image.String())
	if glog.V(4) {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "upload docker image failed")
	}
	glog.V(2).Infof("upload docker image %s finished", image.String())
	return nil
}

func (u *Uploader) Validate(ctx context.Context) error {
	if u.Builder == nil {
		return errors.New("build missing")
	}
	if err := u.Builder.Validate(ctx); err != nil {
		return err
	}
	return nil
}

func (u *Uploader) GetBuilder() world.Builder {
	return u.Builder
}

func (u *Uploader) Satisfied(ctx context.Context) (bool, error) {
	image := u.GetBuilder().GetImage()
	url := fmt.Sprintf("https://index.docker.io/v1/repositories/%s/tags/%s", image.Repository, image.Tag)
	resp, err := http.Get(url)
	if err != nil {
		return false, errors.Wrapf(err, "get %s failed", url)
	}
	if resp.StatusCode/100 != 2 {
		return false, nil
	}
	return true, nil
}
