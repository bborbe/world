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

type Download struct {
	Cluster cluster.Cluster
	Domains []k8s.IngressHost
}

func (d *Download) Applier() world.Applier {
	return nil
}

func (d *Download) Children() []world.Configuration {
	image := docker.Image{
		Registry:   "docker.io",
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
					Ports:         ports,
					CpuLimit:      "250m",
					MemoryLimit:   "25Mi",
					CpuRequest:    "10m",
					MemoryRequest: "10Mi",
					Mounts: []k8s.VolumeMount{
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
					Nfs: k8s.PodNfs{
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

func (d *Download) Validate(ctx context.Context) error {
	glog.V(4).Infof("validate download app ...")
	if err := d.Cluster.Validate(ctx); err != nil {
		return errors.Wrap(err, "validate download app failed")
	}
	if len(d.Domains) == 0 {
		return errors.New("domains empty in download app")
	}
	return nil
}
