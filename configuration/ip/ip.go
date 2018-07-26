package ip

import (
	"context"
	"fmt"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/deploy"
	"github.com/bborbe/world/pkg/docker"
)

type App struct {
	Context world.ClusterContext
	Domains []world.Domain
	Tag     world.Tag
}

func (a *App) image() world.Image {
	return world.Image{
		Registry:   "docker.io",
		Repository: "bborbe/ip",
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
	if len(a.Domains) == 0 {
		return fmt.Errorf("domains empty")
	}
	return nil
}

func (a *App) Deployer() world.Deployer {
	return &deploy.Deployer{
		Image:         a.image(),
		Namespace:     "ip",
		Context:       "netcup",
		Domains:       a.Domains,
		CpuLimit:      "100",
		MemoryLimit:   "50Mi",
		CpuRequest:    "10m",
		MemoryRequest: "10Mi",
		Args:          []world.Arg{"-logtostderr", "-v=2"},
		Port:          8080,
	}
}

func (a *App) Uploader() world.Uploader {
	return &docker.Uploader{
		Image: a.image(),
	}
}

func (a *App) Builder() world.Builder {
	return &docker.GolangBuilder{
		Name:            "ip",
		GitRepo:         "https://github.com/bborbe/ip.git",
		SourceDirectory: "github.com/bborbe/ip",
		Package:         "github.com/bborbe/ip/cmd/ip-server",
		Image:           a.image(),
	}
}
