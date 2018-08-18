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

type Now struct {
	Cluster cluster.Cluster
	Domains []deployer.Domain
	Tag     docker.Tag
}

func (n *Now) Children() []world.Configuration {
	image := docker.Image{
		Registry:   "docker.io",
		Repository: "bborbe/now",
		Tag:        n.Tag,
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
			Context:   n.Cluster.Context,
			Namespace: "now",
		},
		&deployer.DeploymentDeployer{
			Context:   n.Cluster.Context,
			Namespace: "now",
			Name:      "now",
			Requirements: []world.Configuration{
				&build.Now{
					Image: image,
				},
			},
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:          "now",
					Image:         image,
					CpuLimit:      "100m",
					MemoryLimit:   "50Mi",
					CpuRequest:    "10m",
					MemoryRequest: "10Mi",
					Args:          []k8s.Arg{"-logtostderr", "-v=2"},
					Env: []k8s.Env{
						{
							Name:  "PORT",
							Value: "8080",
						},
					},
					Ports: ports,
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   n.Cluster.Context,
			Namespace: "now",
			Name:      "now",
			Ports:     ports,
		},
		&deployer.IngressDeployer{
			Context:   n.Cluster.Context,
			Namespace: "now",
			Name:      "now",
			Domains:   n.Domains,
		},
	}
}

func (n *Now) Applier() world.Applier {
	return nil
}

func (n *Now) Validate(ctx context.Context) error {
	glog.V(4).Infof("validate now app ...")
	if err := n.Cluster.Validate(ctx); err != nil {
		return errors.Wrap(err, "validate now app failed")
	}
	if n.Tag == "" {
		return errors.New("tag missing in now app")
	}
	if len(n.Domains) == 0 {
		return errors.New("domains empty in now app")
	}
	return nil
}
