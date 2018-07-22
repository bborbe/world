package docker

import (
	"context"

	"io/ioutil"
	"os"
	"os/exec"

	"github.com/bborbe/world"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type Builder struct {
	Image   world.Image
	GitRepo world.GitRepo
}

func (b *Builder) Build(ctx context.Context) error {
	glog.V(1).Infof("build docker image ...")

	glog.V(4).Infof("find build dir ...")
	dir, err := ioutil.TempDir("", "build")
	if err != nil {
		return errors.Wrap(err, "find tempdir failed")
	}
	glog.V(4).Infof("found build dir %s", dir)

	{
		glog.V(4).Infof("git clone %s ...", b.GitRepo.String())
		cmd := exec.CommandContext(ctx, "git", "clone", "--branch", b.Image.Tag.String(), "--single-branch", "--depth", "1", b.GitRepo.String(), dir)
		cmd.Dir = dir
		if glog.V(4) {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		}
		if err := cmd.Run(); err != nil {
			return errors.Wrap(err, "clone git repo failed")
		}
	}

	{
		glog.V(4).Infof("docker build %s ...", b.Image.String())
		cmd := exec.CommandContext(ctx, "docker", "build", "--no-cache", "--rm=true", "-t", b.Image.String(), ".")
		cmd.Dir = dir
		if glog.V(4) {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		}
		if err := cmd.Run(); err != nil {
			return errors.Wrap(err, "build docker image failed")
		}
	}

	glog.V(4).Infof("remove build dir %s ...", dir)
	if err := os.RemoveAll(dir); err != nil {
		return errors.Wrap(err, "remove build dir failed")
	}
	return nil
}
