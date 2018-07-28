package mumble

import (
	"context"
	"fmt"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/builder"
	"github.com/bborbe/world/configuration/docker/mumble"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/golang/glog"
)

type App struct {
	Context world.Context
	Tag     world.Tag
}

func (a *App) Required() world.Applier {
	return nil
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

func (a *App) namespaceApplier() world.Applier {
	namespaceBuilder := &builder.NamespaceBuilder{
		Namespace: "mumble",
	}
	return &k8s.Deployer{
		Context: a.Context,
		Data:    namespaceBuilder.Build(),
	}
}
func (a *App) deploymentApplier() world.Applier {
	image := world.Image{
		Registry:   "docker.io",
		Repository: "bborbe/mumble",
		Tag:        a.Tag,
	}
	deploymentBuilder := &builder.DeploymentBuilder{
		Image:         image,
		Namespace:     "mumble",
		Port:          64738,
		HostPort:      64738,
		CpuLimit:      "200m",
		MemoryLimit:   "100Mi",
		CpuRequest:    "100m",
		MemoryRequest: "25Mi",
	}
	return &k8s.Deployer{
		Context: a.Context,
		Data:    deploymentBuilder.Build(),
		Requirements: &mumble.Builder{
			Image: image,
		},
	}
}
func (a *App) serviceApplier() world.Applier {
	serviceBuilder := &builder.ServiceBuilder{
		Namespace: "mumble",
		Port:      64738,
	}
	return &k8s.Deployer{
		Context: a.Context,
		Data:    serviceBuilder.Build(),
	}
}

func (a *App) Apply(ctx context.Context) error {
	glog.V(1).Infof("apply mumble to %s ...", a.Context)
	applier := world.Appliers{
		a.namespaceApplier(),
		a.deploymentApplier(),
		a.serviceApplier(),
	}
	return applier.Apply(ctx)
}

func (a *App) Satisfied(ctx context.Context) (bool, error) {
	return false, nil
}
