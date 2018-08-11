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

type Maven struct {
	Cluster          cluster.Cluster
	Domains          []world.Domain
	MavenRepoVersion world.Tag
}

func (m *Maven) Childs() []world.Configuration {
	nginxImage := world.Image{
		Registry:   "docker.io",
		Repository: "bborbe/nginx-autoindex",
		Tag:        "latest",
	}
	mavenRepoImage := world.Image{
		Registry:   "docker.io",
		Repository: "bborbe/maven-repo",
		Tag:        m.MavenRepoVersion,
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
			Context:   m.Cluster.Context,
			Namespace: "maven",
		},
		&deployer.DeploymentDeployer{
			Context: m.Cluster.Context,
			Requirements: []world.Configuration{
				&docker.NginxAutoindex{
					Image: nginxImage,
				},
			},
			Namespace: "maven",
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:          "nginx",
					Image:         nginxImage,
					Ports:         ports,
					CpuLimit:      "250m",
					MemoryLimit:   "25Mi",
					CpuRequest:    "10m",
					MemoryRequest: "10Mi",
					Mounts: []world.Mount{
						{
							Name:     "maven",
							Target:   "/usr/share/nginx/html",
							ReadOnly: true,
						},
					},
				},
			},
			Volumes: []world.Volume{
				{
					Name:      "maven",
					NfsPath:   "/data/maven",
					NfsServer: m.Cluster.NfsServer,
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   m.Cluster.Context,
			Namespace: "maven",
			Name:      "maven",
			Ports:     ports,
		},
		&deployer.IngressDeployer{
			Context:   m.Cluster.Context,
			Namespace: "maven",
			Domains:   m.Domains,
		},
		&deployer.DeploymentDeployer{
			Context: m.Cluster.Context,
			Requirements: []world.Configuration{
				&docker.Maven{
					Image: mavenRepoImage,
				},
			},
			Namespace: "maven",
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:          "maven",
					Image:         mavenRepoImage,
					CpuLimit:      "100m",
					MemoryLimit:   "50Mi",
					CpuRequest:    "10m",
					MemoryRequest: "10Mi",
					Args:          []world.Arg{"-logtostderr", "-v=1"},
					Ports:         ports,
					Env: []k8s.Env{
						{
							Name:  "ROOT",
							Value: "/data",
						},
					},
					Mounts: []world.Mount{
						{
							Name:   "maven",
							Target: "/data",
						},
					},
				},
			},
			Volumes: []world.Volume{
				{
					Name:      "maven",
					NfsPath:   "/data/maven",
					NfsServer: m.Cluster.NfsServer,
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   m.Cluster.Context,
			Namespace: "maven",
			Name:      "maven-api",
			Ports:     ports,
		},
	}
}

func (m *Maven) Applier() world.Applier {
	return nil
}

func (m *Maven) Validate(ctx context.Context) error {
	glog.V(4).Infof("validate maven app ...")
	if err := m.Cluster.Validate(ctx); err != nil {
		return err
	}
	if m.MavenRepoVersion == "" {
		return fmt.Errorf("tag missing")
	}
	if len(m.Domains) == 0 {
		return fmt.Errorf("domains empty")
	}
	return nil
}
