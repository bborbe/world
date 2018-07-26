package mumble

import (
	"context"

	"fmt"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/deploy"
	"github.com/bborbe/world/pkg/docker"
)

type App struct {
	Context world.ClusterContext
	Tag     world.Tag
}

func (a *App) image() world.Image {
	return world.Image{
		Registry:   "docker.io",
		Repository: "bborbe/mumble",
		Tag:        a.Tag,
	}
}

func (a *App) Validate(ctx context.Context) error {
	if a.Context == "" {
		return fmt.Errorf("context missing")
	}
	if a.Tag == "" {
		return fmt.Errorf("tag missing")
	}
	return nil
}

func (a *App) Deployer() world.Deployer {
	return &deploy.Deployer{
		Image:         a.image(),
		Namespace:     "mumble",
		Context:       "netcup",
		Port:          64738,
		HostPort:      64738,
		CpuLimit:      "200m",
		MemoryLimit:   "100Mi",
		CpuRequest:    "100m",
		MemoryRequest: "25Mi",
	}
}

func (a *App) Uploader() world.Uploader {
	return &docker.Uploader{
		Image: a.image(),
	}
}

func (a *App) Builder() world.Builder {
	return &docker.Builder{
		GitRepo: "https://github.com/bborbe/mumble.git",
		Image:   a.image(),
	}
}
