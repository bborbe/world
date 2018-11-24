// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package docker

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type CloneBuilder struct {
	SourceImage Image
	TargetImage Image
}

func (c *CloneBuilder) Apply(ctx context.Context) error {
	glog.V(1).Infof("docker clone %s ...", c.TargetImage.String())

	glog.V(4).Infof("docker pull %s ...", c.SourceImage.String())
	cmd := createCommand(ctx, "docker", "pull", c.SourceImage.String())
	if glog.V(4) {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "pull docker image failed")
	}

	glog.V(4).Infof("find docker hash for %s", c.SourceImage.Repository.String()+":"+c.SourceImage.Tag.String())
	cmd = createCommand(ctx, "docker", "images", c.SourceImage.Repository.String()+":"+c.SourceImage.Tag.String(), "-q")
	hash := &bytes.Buffer{}
	cmd.Stdout = hash
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "find docker hash failed")
	}

	if hash.Len() == 0 {
		return fmt.Errorf("could not find hash for image: %s", c.SourceImage.Repository.String()+":"+c.SourceImage.Tag.String())
	}

	glog.V(4).Infof("found hash %s", hash.String())

	glog.V(4).Infof("tag image %s with %s", c.SourceImage.String(), c.TargetImage.String())
	cmd = createCommand(ctx, "docker", "tag", strings.TrimSpace(hash.String()), c.TargetImage.String())
	if glog.V(4) {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "tag docker image failed")
	}

	glog.V(1).Infof("docker clone %s finished", c.TargetImage.String())
	return nil
}

func (c *CloneBuilder) Validate(ctx context.Context) error {
	glog.V(4).Infof("validate docker cloner ...")
	if err := c.SourceImage.Validate(ctx); err != nil {
		return errors.Wrap(err, "validate docker cloner failed")
	}
	if err := c.TargetImage.Validate(ctx); err != nil {
		return errors.Wrap(err, "validate docker cloner failed")
	}
	return nil
}

func (c *CloneBuilder) Satisfied(ctx context.Context) (bool, error) {
	return ImageExists(ctx, c.TargetImage)
}

func createCommand(ctx context.Context, name string, arg ...string) *exec.Cmd {
	glog.V(4).Infof("create commmand: %s %s", name, strings.Join(arg, " "))
	return exec.CommandContext(ctx, name, arg...)
}
