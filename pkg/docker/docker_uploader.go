package docker

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"

	"github.com/bborbe/world"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type Uploader struct {
	Image world.Image
}

func (u *Uploader) Upload(ctx context.Context) error {
	glog.V(2).Infof("upload docker image %s ...", u.Image.String())
	cmd := exec.CommandContext(ctx, "docker", "push", u.Image.String())
	if glog.V(4) {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "upload docker image failed")
	}
	glog.V(2).Infof("upload docker image %s finished", u.Image.String())
	return nil
}

func (u *Uploader) Validate(ctx context.Context) error {
	if err := u.Image.Validate(ctx); err != nil {
		return err
	}
	return nil
}

func (u *Uploader) Satisfied(ctx context.Context) (bool, error) {
	url := fmt.Sprintf("https://index.docker.io/v1/repositories/%s/tags/%s", u.Image.Repository, u.Image.Tag)
	resp, err := http.Get(url)
	if err != nil {
		return false, errors.Wrapf(err, "get %s failed", url)
	}
	if resp.StatusCode/100 != 2 {
		return false, nil
	}
	return true, nil
}