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

type Monitoring struct {
	Cluster         cluster.Cluster
	SmtpPassword    deployer.SecretValue
	GitSyncPassword deployer.SecretValue
	GitSyncVersion  docker.Tag
}

func (w *Monitoring) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		w.Cluster,
		w.SmtpPassword,
		w.GitSyncPassword,
		w.GitSyncVersion,
	)
}

func (m *Monitoring) Children() []world.Configuration {
	return []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   m.Cluster.Context,
			Namespace: "monitoring",
		},
		&deployer.SecretDeployer{
			Context:   m.Cluster.Context,
			Namespace: "monitoring",
			Name:      "monitoring",
			Secrets: deployer.Secrets{
				"git-sync-password": m.GitSyncPassword,
				"smtp-password":     m.SmtpPassword,
			},
		},
		//&MonitoringDeployment{
		//
		//},
		//&MonitoringDeployment{
		//
		//},
	}
}

func (m *Monitoring) Applier() (world.Applier, error) {
	return nil, nil
}

type MonitoringDeployment struct {
	Name           string
	Subject        string
	GitRepoUrl     string
	GitSyncVersion docker.Tag
}

func (m *MonitoringDeployment) Children() []world.Configuration {
	gitSyncImage := docker.Image{
		Repository: "bborbe/git-sync",
		Tag:        m.GitSyncVersion,
	}
	return []world.Configuration{
		&deployer.DeploymentDeployer{
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:  "git-sync",
					Image: gitSyncImage,
					Requirement: &build.GitSync{
						Image: gitSyncImage,
					},
					Resources: k8s.PodResources{
						Limits: k8s.Resources{
							Cpu:    "50m",
							Memory: "50Mi",
						},
						Requests: k8s.Resources{
							Cpu:    "10m",
							Memory: "10Mi",
						},
					},
					Args: []k8s.Arg{
						"-logtostderr",
						"-v=4",
					},
					Env: []k8s.Env{
						{
							Name:  "GIT_SYNC_REPO",
							Value: "https://github.com/bborbe/slideshow.git",
						},
						{
							Name:  "GIT_SYNC_DEST",
							Value: "/slideshow",
						},
					},
					Mounts: []k8s.VolumeMount{
						{
							Name: "slideshow",
							Path: "/slideshow",
						},
					},
				},
			},
		},
	}
}

func (m *MonitoringDeployment) Applier() (world.Applier, error) {
	return nil, nil
}
