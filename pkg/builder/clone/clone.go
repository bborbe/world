package clone

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/builder"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type Builder struct {
	SourceImage world.Image
	TargetImage world.Image
}

func (b *Builder) Build(ctx context.Context) error {
	glog.V(1).Infof("docker clone %s ...", b.TargetImage.String())

	glog.V(4).Infof("docker pull %s ...", b.SourceImage.String())
	cmd := exec.CommandContext(ctx, "docker", "pull", b.SourceImage.String())
	if glog.V(4) {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "pull docker image failed")
	}

	glog.V(4).Infof("find docker hash for %s ...", b.SourceImage.String())
	cmd = exec.CommandContext(ctx, "docker", "images", b.SourceImage.Repository.String()+":"+b.SourceImage.Tag.String(), "-q")
	hash := &bytes.Buffer{}
	cmd.Stdout = hash
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "find docker hash failed")
	}
	glog.V(4).Infof("found hash %s", hash.String())

	glog.V(4).Infof("docker tag %s ...", b.SourceImage.String())
	cmd = exec.CommandContext(ctx, "docker", "tag", strings.TrimSpace(hash.String()), b.TargetImage.String())
	if glog.V(4) {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "pull docker image failed")
	}

	glog.V(1).Infof("docker clone %s finished", b.TargetImage.String())
	return nil
}

func (b *Builder) Validate(ctx context.Context) error {
	if err := b.SourceImage.Validate(ctx); err != nil {
		return err
	}
	if err := b.TargetImage.Validate(ctx); err != nil {
		return err
	}
	return nil
}

func (b *Builder) GetImage() world.Image {
	return b.TargetImage
}

func (b *Builder) Satisfied(ctx context.Context) (bool, error) {
	return builder.DockerImageExists(ctx, b.TargetImage)
}
