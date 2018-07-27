package download

import (
	"context"
	"fmt"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/apply"
	"github.com/bborbe/world/pkg/deploy"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/golang/glog"
)

type App struct {
	Context   world.Context
	Domains   []world.Domain
	NfsServer world.MountNfsServer
}

func (a *App) Validate(ctx context.Context) error {
	if a.Context == "" {
		return fmt.Errorf("context missing")
	}
	if a.NfsServer == "" {
		return fmt.Errorf("nfs-server missing")
	}
	if len(a.Domains) == 0 {
		return fmt.Errorf("domains empty")
	}
	return nil
}

func (a *App) Apply(ctx context.Context) error {
	glog.V(1).Infof("apply download to %s ...", a.Context)
	image := world.Image{
		Registry:   "docker.io",
		Repository: "bborbe/nginx-autoindex",
		Tag:        "latest",
	}
	deployer := deploy.Deployer{
		Image:         image,
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
	applier := &apply.Applier{
		Builder: []world.Builder{
			&docker.CloneBuilder{
				SourceImage: world.Image{
					Registry:   "docker.io",
					Repository: "jrelva/nginx-autoindex",
					Tag:        "latest",
				},
				TargetImage: image,
			},
		},
		Uploader: []world.Uploader{
			&docker.Uploader{
				Image: image,
			},
		},
		Deployer: []world.Deployer{
			&k8s.Deployer{
				Context: a.Context,
				Data:    deployer.BuildNamespace(),
			},
			&k8s.Deployer{
				Context: a.Context,
				Data:    deployer.BuildDeployment(),
			},
			&k8s.Deployer{
				Context: a.Context,
				Data:    deployer.BuildService(),
			},
			&k8s.Deployer{
				Context: a.Context,
				Data:    deployer.BuildIngress(),
			},
		},
	}
	return applier.Apply(ctx)
}
