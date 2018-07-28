package ip

import (
	"context"
	"fmt"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/builder"
	"github.com/bborbe/world/configuration/docker/ip"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/golang/glog"
)

type App struct {
	Context world.Context
	Domains []world.Domain
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
	if len(a.Domains) == 0 {
		return fmt.Errorf("domains empty")
	}
	return nil
}

func (a *App) namespaceApplier() world.Applier {
	namespaceBuilder := &builder.NamespaceBuilder{
		Namespace: "ip",
	}
	return &k8s.Deployer{
		Context: a.Context,
		Data:    namespaceBuilder.Build(),
	}
}
func (a *App) deploymentApplier() world.Applier {
	image := world.Image{
		Registry:   "docker.io",
		Repository: "bborbe/ip",
		Tag:        a.Tag,
	}
	deploymentBuilder := &builder.DeploymentBuilder{
		Image:         image,
		Namespace:     "ip",
		CpuLimit:      "100",
		MemoryLimit:   "50Mi",
		CpuRequest:    "10m",
		MemoryRequest: "10Mi",
		Args:          []world.Arg{"-logtostderr", "-v=2"},
		Port:          8080,
	}
	return &k8s.Deployer{
		Context: a.Context,
		Data:    deploymentBuilder.Build(),
		Requirements: &ip.Builder{
			Image: image,
		},
	}
}
func (a *App) serviceApplier() world.Applier {
	serviceBuilder := &builder.ServiceBuilder{
		Namespace: "ip",
		Port:      8080,
	}
	return &k8s.Deployer{
		Context: a.Context,
		Data:    serviceBuilder.Build(),
	}
}
func (a *App) ingressApplier() world.Applier {
	ingressBuilder := &builder.IngressBuilder{
		Namespace: "ip",
		Domains:   a.Domains,
	}
	return &k8s.Deployer{
		Context: a.Context,
		Data:    ingressBuilder.Build(),
	}
}

func (a *App) Apply(ctx context.Context) error {
	glog.V(1).Infof("apply ip to %s ...", a.Context)
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
