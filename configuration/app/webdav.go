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

type Webdav struct {
	Cluster  cluster.Cluster
	Domains  []world.Domain
	Tag      world.Tag
	Password world.SecretValue
}

func (w *Webdav) Childs() []world.Configuration {
	image := world.Image{
		Registry:   "docker.io",
		Repository: "bborbe/webdav",
		Tag:        w.Tag,
	}
	ports := []world.Port{
		{
			Port:     80,
			Name:     "web",
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
			Secrets: world.Secrets{
				"password": w.Password,
			},
		},
		&deployer.DeploymentDeployer{
			Context: w.Cluster.Context,
			Requirements: []world.Configuration{
				&docker.Webdav{
					Image: image,
				},
			},
			Namespace: "webdav",
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:          "webdav",
					Image:         image,
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
					Mounts: []world.Mount{
						{
							Name:   "webdav",
							Target: "/data",
						},
					},
				},
			},
			Volumes: []world.Volume{
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
		return err
	}
	if w.Tag == "" {
		return fmt.Errorf("tag missing")
	}
	if len(w.Domains) == 0 {
		return fmt.Errorf("domains empty")
	}
	if w.Password == nil {
		return fmt.Errorf("password missing")
	}
	if err := w.Password.Validate(ctx); err != nil {
		return err
	}
	return nil
}
