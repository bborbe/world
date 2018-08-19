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

type Webdav struct {
	Cluster  cluster.Cluster
	Domains  []deployer.Domain
	Tag      docker.Tag
	Password deployer.SecretValue
}

func (w *Webdav) Children() []world.Configuration {
	image := docker.Image{
		Registry:   "docker.io",
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
					CpuLimit:      "50m",
					MemoryLimit:   "50Mi",
					CpuRequest:    "10m",
					MemoryRequest: "10Mi",
					Ports:         ports,
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
					Mounts: []deployer.Mount{
						{
							Name:   "webdav",
							Target: "/data",
						},
					},
				},
			},
			Volumes: []deployer.Volume{
				{
					Name:      "webdav",
					NfsPath:   "/data/webdav",
					NfsServer: w.Cluster.NfsServer,
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
			Domains:   w.Domains,
		},
	}
}

func (w *Webdav) Applier() world.Applier {
	return nil
}

func (w *Webdav) Validate(ctx context.Context) error {
	glog.V(4).Infof("validate webdav app ...")
	if err := w.Cluster.Validate(ctx); err != nil {
		return errors.Wrap(err, "validate webdav app failed")
	}
	if w.Tag == "" {
		return errors.New("tag missing in webdav app")
	}
	if len(w.Domains) == 0 {
		return errors.New("domains empty in webdav app")
	}
	if w.Password == nil {
		return errors.New("password missing in webdav app")
	}
	if err := w.Password.Validate(ctx); err != nil {
		return errors.Wrap(err, "validate webdav app failed")
	}
	return nil
}
