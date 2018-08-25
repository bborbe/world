package app

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
)

type Download struct {
	Cluster cluster.Cluster
	Domains k8s.IngressHosts
}

func (t *Download) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Cluster,
		t.Domains,
	)
}

func (d *Download) Applier() (world.Applier, error) {
	return nil, nil
}

func (d *Download) Children() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/nginx-autoindex",
		Tag:        "latest",
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
			Context:   d.Cluster.Context,
			Namespace: "download",
		},
		&deployer.DeploymentDeployer{
			Context:   d.Cluster.Context,
			Namespace: "download",
			Name:      "download",
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:  "nginx",
					Image: image,
					Requirement: &build.NginxAutoindex{
						Image: image,
					},
					Ports: ports,
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "250m",
							Memory: "25Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "10m",
							Memory: "10Mi",
						},
					},
					Mounts: []k8s.ContainerMount{
						{
							Name:     "download",
							Path:     "/usr/share/nginx/html",
							ReadOnly: true,
						},
					},
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name: "download",
					Nfs: k8s.PodVolumeNfs{
						Path:   "/data/download",
						Server: d.Cluster.NfsServer,
					},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   d.Cluster.Context,
			Namespace: "download",
			Name:      "download",
			Ports:     ports,
		},
		&deployer.IngressDeployer{
			Context:   d.Cluster.Context,
			Namespace: "download",
			Name:      "download",
			Port:      "http",
			Domains:   d.Domains,
		},
	}
}
