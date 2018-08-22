package docker

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type Builder struct {
	Image     Image
	GitRepo   GitRepo
	GitBranch GitBranch
	BuildArgs BuildArgs
}

func (b *Builder) Apply(ctx context.Context) error {
	glog.V(1).Infof("build docker image %s ...", b.Image.String())

	glog.V(4).Infof("find build dir ...")
	dir, err := ioutil.TempDir("", "build")
	if err != nil {
		return errors.Wrap(err, "find tempdir failed")
	}
	glog.V(4).Infof("found build dir %s", dir)

	{
		glog.V(4).Infof("git clone %s ...", b.GitRepo.String())
		cmd := exec.CommandContext(ctx, "git", "clone", "--branch", b.GitBranch.String(), "--single-branch", "--depth", "1", b.GitRepo.String(), dir)
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
		var args []string
		args = append(args, "build")
		args = append(args, "--no-cache", "--rm=true")
		for k, v := range b.BuildArgs {
			args = append(args, "--build-arg", fmt.Sprintf("%s=%s", k, v))
		}
		args = append(args, "-t", b.Image.String(), ".")
		glog.V(4).Infof("docker build %s ...", b.Image.String())
		cmd := exec.CommandContext(ctx, "docker", args...)
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
	glog.V(1).Infof("build docker image %s finished", b.Image.String())
	return nil
}

func (b *Builder) Validate(ctx context.Context) error {
	glog.V(4).Infof("validate docker builder ...")
	if err := b.Image.Validate(ctx); err != nil {
		return errors.Wrap(err, "validate docker builder failed")
	}
	if b.GitRepo == "" {
		return errors.New("git repo missing")
	}
	if b.GitBranch == "" {
		return errors.New("git branch missing")
	}
	return nil
}

func (b *Builder) Satisfied(ctx context.Context) (bool, error) {
	return ImageExists(ctx, b.Image)
}
