package download

import (
	"context"
	"fmt"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/builder"
	"github.com/bborbe/world/configuration/docker/download"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/golang/glog"
)

var Namespace = world.Namespace("download")

type App struct {
	Context   world.Context
	Domains   []world.Domain
	NfsServer world.MountNfsServer
}

func (a *App) Required() world.Applier {
	return nil
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

func (a *App) namespaceApplier() world.Applier {
	namespaceBuilder := &builder.NamespaceBuilder{
		Namespace: Namespace,
	}
	return &k8s.Deployer{
		Context: a.Context,
		Data:    namespaceBuilder.Build(),
	}
}
func (a *App) deploymentApplier() world.Applier {
	image := world.Image{
		Registry:   "docker.io",
		Repository: "bborbe/nginx-autoindex",
		Tag:        "latest",
	}
	deploymentBuilder := &builder.DeploymentBuilder{
		Image:         image,
		Namespace:     "download",
		Port:          80,
		CpuLimit:      "250m",
		MemoryLimit:   "25Mi",
		CpuRequest:    "10m",
		MemoryRequest: "10Mi",
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
	return &k8s.Deployer{
		Context: a.Context,
		Data:    deploymentBuilder.Build(),
		Requirements: &download.Builder{
			Image: image,
		},
	}
}
func (a *App) serviceApplier() world.Applier {
	serviceBuilder := &builder.ServiceBuilder{
		Namespace: "download",
		Port:      80,
	}
	return &k8s.Deployer{
		Context: a.Context,
		Data:    serviceBuilder.Build(),
	}
}
func (a *App) ingressApplier() world.Applier {
	ingressBuilder := &builder.IngressBuilder{
		Namespace: "download",
		Domains:   a.Domains,
	}
	return &k8s.Deployer{
		Context: a.Context,
		Data:    ingressBuilder.Build(),
	}
}

func (a *App) Apply(ctx context.Context) error {
	glog.V(1).Infof("apply download to %s ...", a.Context)
	applier := world.Appliers{
		a.namespaceApplier(),
		a.deploymentApplier(),
		a.serviceApplier(),
		a.ingressApplier(),
	}
	return applier.Apply(ctx)
}

func (a *App) Satisfied(ctx context.Context) (bool, error) {
	return false, nil
}
