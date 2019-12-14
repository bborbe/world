// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package docker

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"time"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

func ImageExists(ctx context.Context, image Image) (bool, error) {
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

type SourceDirectory string

func (s SourceDirectory) String() string {
	return string(s)
}

type GitRepo string

func (g GitRepo) String() string {
	return string(g)
}

type GitBranch string

func (g GitBranch) String() string {
	return string(g)
}

func (g GitBranch) Validate(ctx context.Context) error {
	if g == "" {
		return errors.New("GitBranch empty")
	}
	return nil
}

type BuildArgs map[string]string

type Registry string

func (r Registry) String() string {
	return string(r)
}

type Repositories []Repository

func (r Repositories) Validate(ctx context.Context) error {
	for _, r := range r {
		if err := r.Validate(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (r Repositories) String() string {
	buf := &bytes.Buffer{}
	for i, repo := range r {
		if i != 0 {
			buf.WriteString(",")
		}
		buf.WriteString(repo.String())
	}
	return buf.String()
}

type Repository string

func (r Repository) String() string {
	return string(r)
}

func (r Repository) Validate(ctx context.Context) error {
	if r == "" {
		return errors.New("Repository empty")
	}
	return nil
}

func TagWithTime(version string, now time.Time) Tag {
	return Tag(version + "-" + now.Format("200601"))
}

type Tag string

func (t Tag) String() string {
	return string(t)
}

func (t Tag) Validate(ctx context.Context) error {
	if t == "" {
		return errors.New("Tag empty")
	}
	return nil
}

type Image struct {
	Repository Repository
	Tag        Tag
}

func (i Image) String() string {
	return i.Repository.String() + ":" + i.Tag.String()
}

func (i Image) Validate(ctx context.Context) error {
	glog.V(4).Infof("validate image ...")
	if i.Tag == "" {
		return errors.New("Tag missing")
	}
	if i.Repository == "" {
		return errors.New("Repository missing")
	}
	return nil
}
