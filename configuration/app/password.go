package app

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type Password struct {
	Cluster cluster.Cluster
	Domains []deployer.Domain
	Tag     docker.Tag
}

func (p *Password) Children() []world.Configuration {
	image := docker.Image{
		Registry:   "docker.io",
		Repository: "bborbe/password",
		Tag:        p.Tag,
	}
	ports := []deployer.Port{
		{
			Port:     8080,
			Name:     "http",
			Protocol: "TCP",
		},
	}
	return []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   p.Cluster.Context,
			Namespace: "password",
		},
		&deployer.DeploymentDeployer{
			Context:   p.Cluster.Context,
			Namespace: "password",
			Name:      "password",
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:  "password",
					Image: image,
					Requirement: &build.Password{
						Image: image,
					},
					CpuLimit:      "100m",
					MemoryLimit:   "50Mi",
					CpuRequest:    "10m",
					MemoryRequest: "10Mi",
					Args:          []k8s.Arg{"-logtostderr", "-v=2"},
					Ports:         ports,
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   p.Cluster.Context,
			Namespace: "password",
			Name:      "password",
			Ports:     ports,
		},
		&deployer.IngressDeployer{
			Context:   p.Cluster.Context,
			Namespace: "password",
			Name:      "password",
			Domains:   p.Domains,
		},
	}
}

func (p *Password) Applier() world.Applier {
	return nil
}

func (p *Password) Validate(ctx context.Context) error {
	glog.V(4).Infof("validate password app ...")
	if err := p.Cluster.Validate(ctx); err != nil {
		return errors.Wrap(err, "validate password app failed")
	}
	if p.Tag == "" {
		return errors.New("tag missing in password app")
	}
	if len(p.Domains) == 0 {
		return errors.New("domains empty in password app")
	}
	return nil
}
