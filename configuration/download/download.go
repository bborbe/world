package download

import (
	"context"
	"fmt"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/deploy"
	"github.com/bborbe/world/pkg/docker"
)

type App struct {
	Context   world.ClusterContext
	Domains   []world.Domain
	Tag       world.Tag
	NfsServer world.MountNfsServer
}

func (a *App) image() world.Image {
	return world.Image{
		Registry:   "docker.io",
		Repository: "bborbe/nginx-autoindex",
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
		Namespace:     "download",
		Context:       "netcup",
		Port:          80,
		CpuLimit:      "250m",
		MemoryLimit:   "25Mi",
		CpuRequest:    "10m",
		MemoryRequest: "10Mi",
		Domains:       a.Domains,
		Mounts: []world.Mount{
			{
				Name:      "download",
				Target:    "/usr/share/nginx/html",
				ReadOnly:  true,
				NfsPath:   "/data/download",
				NfsServer: a.NfsServer,
			},
		},
	}
}

func (a *App) Uploader() world.Uploader {
	return &docker.Uploader{
		Image: a.image(),
	}
}

func (a *App) Builder() world.Builder {
	return &docker.CloneBuilder{
		SourceImage: world.Image{
			Registry:   "docker.io",
			Repository: "jrelva/nginx-autoindex",
			Tag:        a.Tag,
		},
		TargetImage: a.image(),
	}
}
