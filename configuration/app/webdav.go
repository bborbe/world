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

type Webdav struct {
	Cluster  cluster.Cluster
	Domains  k8s.IngressHosts
	Tag      docker.Tag
	Password deployer.SecretValue
}

func (t *Webdav) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Cluster,
		t.Domains,
		t.Tag,
		t.Password,
	)
}

func (w *Webdav) Children() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/webdav",
		Tag:        w.Tag,
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
			Context:   w.Cluster.Context,
			Namespace: "webdav",
		},
		&deployer.SecretDeployer{
			Context:   w.Cluster.Context,
			Namespace: "webdav",
			Name:      "webdav",
			Secrets: deployer.Secrets{
				"password": w.Password,
			},
		},
		&deployer.DeploymentDeployer{
			Context:   w.Cluster.Context,
			Namespace: "webdav",
			Name:      "webdav",
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:  "webdav",
					Image: image,
					Requirement: &build.Webdav{
						Image: image,
					},
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "50m",
							Memory: "50Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "10m",
							Memory: "10Mi",
						},
					},
					Ports: ports,
					Env: []k8s.Env{
						{
							Name:  "WEBDAV_USERNAME",
							Value: "bborbe",
						},
						{
							Name: "WEBDAV_PASSWORD",
							ValueFrom: k8s.ValueFrom{
								SecretKeyRef: k8s.SecretKeyRef{
									Key:  "password",
									Name: "webdav",
								},
							},
						},
					},
					Mounts: []k8s.ContainerMount{
						{
							Name: "webdav",
							Path: "/data",
						},
					},
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name: "webdav",
					Nfs: k8s.PodVolumeNfs{
						Path:   "/data/webdav",
						Server: w.Cluster.NfsServer,
					},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   w.Cluster.Context,
			Namespace: "webdav",
			Name:      "webdav",
			Ports:     ports,
		},
		&deployer.IngressDeployer{
			Context:   w.Cluster.Context,
			Namespace: "webdav",
			Name:      "webdav",
			Port:      "http",
			Domains:   w.Domains,
		},
	}
}

func (w *Webdav) Applier() (world.Applier, error) {
	return nil, nil
}
