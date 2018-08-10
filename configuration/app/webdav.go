package app

import (
	"context"
	"fmt"

	"github.com/bborbe/teamvault-utils/connector"
	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/configuration/docker"
	"github.com/bborbe/world/pkg/k8s"
)

type Webdav struct {
	Context            world.Context
	Domains            []world.Domain
	Tag                world.Tag
	NfsServer          world.MountNfsServer
	TeamvaultConnector connector.Connector
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
			Context:   w.Context,
			Namespace: "webdav",
		},
		&deployer.SecretDeployer{
			Context:   w.Context,
			Namespace: "webdav",
			Secrets: world.Secrets{
				"password": &world.SecretFromTeamvault{
					TeamvaultConnector: w.TeamvaultConnector,
					TeamvaultKey:       "VOzvAO",
				},
			},
		},
		&deployer.DeploymentDeployer{
			Context: w.Context,
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
					NfsServer: w.NfsServer,
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   w.Context,
			Namespace: "webdav",
			Ports:     ports,
		},
		&deployer.IngressDeployer{
			Context:   w.Context,
			Namespace: "webdav",
			Domains:   w.Domains,
		},
	}
}

func (w *Webdav) Applier() world.Applier {
	return nil
}

func (w *Webdav) Validate(ctx context.Context) error {
	if w.Context == "" {
		return fmt.Errorf("context missing")
	}
	if w.Tag == "" {
		return fmt.Errorf("tag missing")
	}
	if len(w.Domains) == 0 {
		return fmt.Errorf("domains empty")
	}
	if w.NfsServer == "" {
		return fmt.Errorf("nfs-server missing")
	}
	if w.TeamvaultConnector == nil {
		return fmt.Errorf("teamvault-connector missing")
	}
	return nil
}
