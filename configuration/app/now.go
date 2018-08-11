package app

import (
	"context"
	"fmt"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/configuration/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/golang/glog"
)

type Now struct {
	Cluster cluster.Cluster
	Domains []world.Domain
	Tag     world.Tag
}

func (n *Now) Childs() []world.Configuration {
	image := world.Image{
		Registry:   "docker.io",
		Repository: "bborbe/now",
		Tag:        n.Tag,
	}
	ports := []world.Port{
		{
			Port:     8080,
			Name:     "web",
			Protocol: "TCP",
		},
	}
	return []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   n.Cluster.Context,
			Namespace: "now",
		},
		&deployer.DeploymentDeployer{
			Context: n.Cluster.Context,
			Requirements: []world.Configuration{
				&docker.Now{
					Image: image,
				},
			},
			Namespace: "now",
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:          "now",
					Image:         image,
					CpuLimit:      "100m",
					MemoryLimit:   "50Mi",
					CpuRequest:    "10m",
					MemoryRequest: "10Mi",
					Args:          []world.Arg{"-logtostderr", "-v=2"},
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
		return err
	}
	if n.Tag == "" {
		return fmt.Errorf("tag missing")
	}
	if len(n.Domains) == 0 {
		return fmt.Errorf("domains empty")
	}
	return nil
}
