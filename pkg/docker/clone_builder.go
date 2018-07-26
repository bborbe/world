package docker

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"

	"github.com/bborbe/world"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type CloneBuilder struct {
	SourceImage world.Image
	TargetImage world.Image
}

func (c *CloneBuilder) Build(ctx context.Context) error {
	glog.V(1).Infof("docker clone %s ...", c.TargetImage.String())

	glog.V(4).Infof("docker pull %s ...", c.SourceImage.String())
	cmd := exec.CommandContext(ctx, "docker", "pull", c.SourceImage.String())
	if glog.V(4) {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "pull docker image failed")
	}

	glog.V(4).Infof("find docker hash for %s ...", c.SourceImage.String())
	cmd = exec.CommandContext(ctx, "docker", "images", c.SourceImage.Repository.String()+":"+c.SourceImage.Tag.String(), "-q")
	hash := &bytes.Buffer{}
	cmd.Stdout = hash
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "find docker hash failed")
	}
	glog.V(4).Infof("found hash %s", hash.String())

	glog.V(4).Infof("docker tag %s ...", c.SourceImage.String())
	cmd = exec.CommandContext(ctx, "docker", "tag", strings.TrimSpace(hash.String()), c.TargetImage.String())
	if glog.V(4) {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "pull docker image failed")
	}

	glog.V(1).Infof("docker clone %s finished", c.TargetImage.String())
	return nil
}

func (c *CloneBuilder) Validate(ctx context.Context) error {
	if err := c.SourceImage.Validate(ctx); err != nil {
		return err
	}
	if err := c.TargetImage.Validate(ctx); err != nil {
		return err
	}
	return nil
}

func (c *CloneBuilder) GetImage() world.Image {
	return c.TargetImage
}

func (c *CloneBuilder) Satisfied(ctx context.Context) (bool, error) {
	return ImageExists(ctx, c.TargetImage)
}
