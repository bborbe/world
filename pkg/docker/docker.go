package docker

import (
	"context"
	"os"
	"os/exec"

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

type BuildArgs map[string]string

type Registry string

func (r Registry) String() string {
	return string(r)
}

type Repository string

func (i Repository) String() string {
	return string(i)
}

type Tag string

func (v Tag) String() string {
	return string(v)
}

type Image struct {
	Repository Repository
	Registry   Registry
	Tag        Tag
}

func (i Image) String() string {
	return i.Registry.String() + "/" + i.Repository.String() + ":" + i.Tag.String()
}

func (b *Image) Validate(ctx context.Context) error {
	glog.V(4).Infof("validate image ...")
	if b.Tag == "" {
		return errors.New("Tag missing")
	}
	if b.Registry == "" {
		return errors.New("Registry missing")
	}
	if b.Repository == "" {
		return errors.New("Repository missing")
	}
	return nil
}
