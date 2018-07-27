package mumble

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
	Context world.Context
	Tag     world.Tag
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

func (a *App) Apply(ctx context.Context) error {
	glog.V(1).Infof("apply mumble to %s ...", a.Context)
	image := world.Image{
		Registry:   "docker.io",
		Repository: "bborbe/mumble",
		Tag:        a.Tag,
	}
	deployer := &deploy.Deployer{
		Image:         image,
		Namespace:     "mumble",
		Context:       "netcup",
		Port:          64738,
		HostPort:      64738,
		CpuLimit:      "200m",
		MemoryLimit:   "100Mi",
		CpuRequest:    "100m",
		MemoryRequest: "25Mi",
	}
	applier := &apply.Applier{
		Builder: []world.Builder{
			&docker.Builder{
				GitRepo: "https://github.com/bborbe/mumble.git",
				Image:   image,
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
		},
	}
	return applier.Apply(ctx)
}
