package app

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type Mumble struct {
	Cluster cluster.Cluster
	Tag     docker.Tag
}

func (m *Mumble) Children() []world.Configuration {
	image := docker.Image{
		Registry:   "docker.io",
		Repository: "bborbe/mumble",
		Tag:        m.Tag,
	}
	ports := []deployer.Port{
		{
			Port:     64738,
			HostPort: 64738,
			Protocol: "TCP",
		},
	}
	return []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   m.Cluster.Context,
			Namespace: "mumble",
		},
		&deployer.DeploymentDeployer{
			Context:   m.Cluster.Context,
			Namespace: "mumble",
			Name:      "mumble",
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:  "mumble",
					Image: image,
					Requirement: &build.Mumble{
						Image: image,
					},
					Ports:         ports,
					CpuLimit:      "200m",
					MemoryLimit:   "100Mi",
					CpuRequest:    "100m",
					MemoryRequest: "25Mi",
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   m.Cluster.Context,
			Namespace: "mumble",
			Name:      "mumble",
			Ports:     ports,
		},
	}
}

func (m *Mumble) Applier() world.Applier {
	return nil
}

func (m *Mumble) Validate(ctx context.Context) error {
	glog.V(4).Infof("validate mumble app ...")
	if err := m.Cluster.Validate(ctx); err != nil {
		return errors.Wrap(err, "validate mumble app failed")
	}
	if m.Tag == "" {
		return errors.New("tag missing in mumble app")
	}
	return nil
}
