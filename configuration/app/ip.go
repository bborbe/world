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

type Ip struct {
	Cluster cluster.Cluster
	Domains []deployer.Domain
	Tag     docker.Tag
}

func (i *Ip) Applier() world.Applier {
	return nil
}

func (i *Ip) Children() []world.Configuration {
	image := docker.Image{
		Registry:   "docker.io",
		Repository: "bborbe/ip",
		Tag:        i.Tag,
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
			Context:   i.Cluster.Context,
			Namespace: "ip",
		},
		&deployer.DeploymentDeployer{
			Context:   i.Cluster.Context,
			Namespace: "ip",
			Name:      "ip",
			Requirements: []world.Configuration{
				&build.Ip{
					Image: image,
				},
			},
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:          "ip",
					Image:         image,
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
			Context:   i.Cluster.Context,
			Namespace: "ip",
			Name:      "ip",
			Ports:     ports,
		},
		&deployer.IngressDeployer{
			Context:   i.Cluster.Context,
			Namespace: "ip",
			Name:      "ip",
			Domains:   i.Domains,
		},
	}
}

func (i *Ip) Validate(ctx context.Context) error {
	glog.V(4).Infof("validate ip app ...")
	if err := i.Cluster.Validate(ctx); err != nil {
		return errors.Wrap(err, "validate ip app failed")
	}
	if i.Tag == "" {
		return errors.New("tag missing")
	}
	if len(i.Domains) == 0 {
		return errors.New("domains empty")
	}
	return nil
}
