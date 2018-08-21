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

type HelloWorld struct {
	Cluster cluster.Cluster
	Domains []k8s.IngressHost
	Tag     docker.Tag
}

func (h *HelloWorld) Applier() world.Applier {
	return nil
}

func (h *HelloWorld) Children() []world.Configuration {
	image := docker.Image{
		Registry:   "docker.io",
		Repository: "bborbe/hello-world",
		Tag:        h.Tag,
	}
	ports := []deployer.Port{
		{
			Port:     80,
			Name:     "http",
			Protocol: "TCP",
		},
	}
	return []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   h.Cluster.Context,
			Namespace: "hello-world",
		},
		&deployer.DeploymentDeployer{
			Context:   h.Cluster.Context,
			Namespace: "hello-world",
			Name:      "hello-world",
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:  "hello-world",
					Image: image,
					Requirement: &build.HelloWorld{
						Image: image,
					},
					CpuLimit:      "100m",
					MemoryLimit:   "50Mi",
					CpuRequest:    "10m",
					MemoryRequest: "10Mi",
					Ports:         ports,
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   h.Cluster.Context,
			Namespace: "hello-world",
			Name:      "hello-world",
			Ports:     ports,
		},
		&deployer.IngressDeployer{
			Context:   h.Cluster.Context,
			Namespace: "hello-world",
			Name:      "hello-world",
			Port:      "http",
			Domains:   h.Domains,
		},
	}
}

func (h *HelloWorld) Validate(ctx context.Context) error {
	glog.V(4).Infof("validate hello-world app ...")
	if err := h.Cluster.Validate(ctx); err != nil {
		return errors.Wrap(err, "validate hello world app failed")
	}
	if len(h.Domains) == 0 {
		return errors.New("Domains empty")
	}
	if h.Tag == "" {
		return errors.New("Tag missing")
	}
	return nil
}
