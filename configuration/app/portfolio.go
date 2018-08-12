package app

import (
	"context"

	"strconv"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type Portfolio struct {
	Cluster              cluster.Cluster
	Domains              []world.Domain
	OverlayServerVersion docker.Tag
	GitSyncVersion       docker.Tag
	GitSyncPassword      world.SecretValue
}

func (p *Portfolio) Childs() []world.Configuration {
	port := 8080
	overlayServerImage := docker.Image{
		Registry:   "docker.io",
		Repository: "bborbe/portfolio",
		Tag:        p.OverlayServerVersion,
	}
	gitSyncImage := docker.Image{
		Registry:   "docker.io",
		Repository: "bborbe/git-sync",
		Tag:        p.GitSyncVersion,
	}
	ports := []world.Port{
		{
			Port:     port,
			Name:     "web",
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
			Secrets: world.Secrets{
				"git-sync-password": p.GitSyncPassword,
			},
		},
		&deployer.DeploymentDeployer{
			Context:   p.Cluster.Context,
			Namespace: "portfolio",
			Name:      "portfolio",
			Requirements: []world.Configuration{
				&build.OverlayWebserver{
					Image: overlayServerImage,
				},
				&build.GitSync{
					Image: gitSyncImage,
				},
			},
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:          "portfolio",
					Image:         overlayServerImage,
					CpuLimit:      "50m",
					MemoryLimit:   "50Mi",
					CpuRequest:    "10m",
					MemoryRequest: "10Mi",
					Args:          []world.Arg{"-logtostderr", "-v=1"},
					Ports:         ports,
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
					Mounts: []world.Mount{
						{
							Name:     "portfolio",
							Target:   "/portfolio",
							ReadOnly: true,
						},
						{
							Name:     "overlay",
							Target:   "/overlay",
							ReadOnly: true,
						},
					},
				},
				{
					Name:          "git-sync-portfolio",
					Image:         gitSyncImage,
					CpuLimit:      "50m",
					MemoryLimit:   "50Mi",
					CpuRequest:    "10m",
					MemoryRequest: "10Mi",
					Args: []world.Arg{
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
					Mounts: []world.Mount{
						{
							Name:   "portfolio",
							Target: "/portfolio",
						},
					},
				},
				{
					Name:          "git-sync-overlay",
					Image:         gitSyncImage,
					CpuLimit:      "50m",
					MemoryLimit:   "50Mi",
					CpuRequest:    "10m",
					MemoryRequest: "10Mi",
					Args: []world.Arg{
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
					Mounts: []world.Mount{
						{
							Name:   "overlay",
							Target: "/overlay",
						},
					},
				},
			},
			Volumes: []world.Volume{
				{
					Name:     "portfolio",
					EmptyDir: true,
				},
				{
					Name:     "overlay",
					EmptyDir: true,
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
			Domains:   p.Domains,
		},
	}
}

func (p *Portfolio) Applier() world.Applier {
	return nil
}

func (p *Portfolio) Validate(ctx context.Context) error {
	glog.V(4).Infof("validate portfolio app ...")
	if err := p.Cluster.Validate(ctx); err != nil {
		return errors.Wrap(err, "validate portfolio app failed")
	}
	if p.OverlayServerVersion == "" {
		return errors.New("tag missing in portfolio app")
	}
	if len(p.Domains) == 0 {
		return errors.New("domains empty in portfolio app")
	}
	if p.GitSyncVersion == "" {
		return errors.New("git-sync-version missing in portfolio app")
	}
	if p.GitSyncPassword == nil {
		return errors.New("git-sync-password missing in portfolio app")
	}
	if err := p.GitSyncPassword.Validate(ctx); err != nil {
		return errors.Wrap(err, "validate portfolio app failed")
	}
	return nil
}
