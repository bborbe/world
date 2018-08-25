package app

import (
	"strconv"

	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
)

type Portfolio struct {
	Cluster              cluster.Cluster
	Domains              k8s.IngressHosts
	OverlayServerVersion docker.Tag
	GitSyncVersion       docker.Tag
	GitSyncPassword      deployer.SecretValue
}

func (t *Portfolio) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Cluster,
		t.Domains,
		t.OverlayServerVersion,
		t.GitSyncVersion,
		t.GitSyncPassword,
	)
}

func (p *Portfolio) Children() []world.Configuration {
	port := 8080
	overlayServerImage := docker.Image{
		Repository: "bborbe/portfolio",
		Tag:        p.OverlayServerVersion,
	}
	gitSyncImage := docker.Image{
		Repository: "bborbe/git-sync",
		Tag:        p.GitSyncVersion,
	}
	ports := []deployer.Port{
		{
			Port:     port,
			Name:     "http",
			Protocol: "TCP",
		},
	}
	return []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   p.Cluster.Context,
			Namespace: "portfolio",
		},
		&deployer.SecretDeployer{
			Context:   p.Cluster.Context,
			Namespace: "portfolio",
			Name:      "portfolio",
			Secrets: deployer.Secrets{
				"git-sync-password": p.GitSyncPassword,
			},
		},
		&deployer.DeploymentDeployer{
			Context:   p.Cluster.Context,
			Namespace: "portfolio",
			Name:      "portfolio",
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:  "portfolio",
					Image: overlayServerImage,
					Requirement: &build.OverlayWebserver{
						Image: overlayServerImage,
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
					Args:  []k8s.Arg{"-logtostderr", "-v=1"},
					Ports: ports,
					Env: []k8s.Env{
						{
							Name:  "PORT",
							Value: strconv.Itoa(port),
						},
						{
							Name:  "ROOT",
							Value: "/portfolio/files",
						},
						{
							Name:  "OVERLAYS",
							Value: "/overlay",
						},
					},
					Mounts: []k8s.VolumeMount{
						{
							Name:     "portfolio",
							Path:     "/portfolio",
							ReadOnly: true,
						},
						{
							Name:     "overlay",
							Path:     "/overlay",
							ReadOnly: true,
						},
					},
				},
				{
					Name:  "git-sync-portfolio",
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
						"-v=1",
					},
					Env: []k8s.Env{
						{
							Name:  "GIT_SYNC_REPO",
							Value: "https://github.com/bborbe/portfolio.git",
						},
						{
							Name:  "GIT_SYNC_DEST",
							Value: "/portfolio",
						},
					},
					Mounts: []k8s.VolumeMount{
						{
							Name: "portfolio",
							Path: "/portfolio",
						},
					},
				},
				{
					Name:  "git-sync-overlay",
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
						"-v=1",
					},
					Env: []k8s.Env{
						{
							Name:  "GIT_SYNC_REPO",
							Value: "https://bborbereadonly@bitbucket.org/bborbe/benjaminborbe_portfolio.git",
						},
						{
							Name:  "GIT_SYNC_DEST",
							Value: "/overlay",
						},
						{
							Name:  "GIT_SYNC_USERNAME",
							Value: "bborbereadonly",
						},
						{
							Name: "GIT_SYNC_PASSWORD",
							ValueFrom: k8s.ValueFrom{
								SecretKeyRef: k8s.SecretKeyRef{
									Key:  "git-sync-password",
									Name: "portfolio",
								},
							},
						},
					},
					Mounts: []k8s.VolumeMount{
						{
							Name: "overlay",
							Path: "/overlay",
						},
					},
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name:     "portfolio",
					EmptyDir: &k8s.PodVolumeEmptyDir{},
				},
				{
					Name:     "overlay",
					EmptyDir: &k8s.PodVolumeEmptyDir{},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   p.Cluster.Context,
			Namespace: "portfolio",
			Name:      "portfolio",
			Ports:     ports,
		},
		&deployer.IngressDeployer{
			Context:   p.Cluster.Context,
			Namespace: "portfolio",
			Name:      "portfolio",
			Port:      "http",
			Domains:   p.Domains,
		},
	}
}

func (p *Portfolio) Applier() (world.Applier, error) {
	return nil, nil
}
