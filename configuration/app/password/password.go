package password

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
	Domains []world.Domain
	Tag     world.Tag
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

func (a *App) Apply(ctx context.Context) error {
	glog.V(1).Infof("apply password to %s ...", a.Context)
	image := world.Image{
		Registry:   "docker.io",
		Repository: "bborbe/password",
		Tag:        a.Tag,
	}
	deployer := &deploy.Deployer{
		Image:         image,
		Namespace:     "password",
		Context:       "netcup",
		Domains:       a.Domains,
		CpuLimit:      "100",
		MemoryLimit:   "50Mi",
		CpuRequest:    "10m",
		MemoryRequest: "10Mi",
		Args:          []world.Arg{"-logtostderr", "-v=2"},
		Port:          8080,
	}
	applier := &apply.Applier{
		Builder: []world.Builder{
			&docker.GolangBuilder{
				Name:            "password",
				GitRepo:         "https://github.com/bborbe/password.git",
				SourceDirectory: "github.com/bborbe/password",
				Package:         "github.com/bborbe/password/cmd/password-server",
				Image:           image,
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
