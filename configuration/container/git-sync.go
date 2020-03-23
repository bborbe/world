// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package container

import (
	"context"
	"fmt"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
	"github.com/pkg/errors"
)

type GitRepoUrl string

func (g GitRepoUrl) String() string {
	return string(g)
}

func (g GitRepoUrl) Validate(ctx context.Context) error {
	if g == "" {
		return errors.New("GitRepoUrl missing")
	}
	return nil
}

type GitSync struct {
	MountName                 k8s.MountName
	GitRepoUrl                GitRepoUrl
	GitSyncUsername           string
	GitSyncPasswordSecretName string
	GitSyncPasswordSecretPath string
}

func (g *GitSync) image() docker.Image {
	return docker.Image{
		Repository: "bborbe/git-sync",
		Tag:        "1.3.0",
	}
}

func (g *GitSync) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		g.MountName,
		g.GitRepoUrl,
	)
}

func (g *GitSync) Requirements() []world.Configuration {
	return []world.Configuration{
		&build.GitSync{
			Image: g.image(),
		},
	}
}

func (g *GitSync) Container() k8s.Container {
	envs := []k8s.Env{
		{
			Name:  "GIT_SYNC_REPO",
			Value: g.GitRepoUrl.String(),
		},
		{
			Name:  "GIT_SYNC_DEST",
			Value: "/target",
		},
	}
	if g.GitSyncUsername != "" {
		envs = append(envs, k8s.Env{
			Name:  "GIT_SYNC_USERNAME",
			Value: g.GitSyncUsername,
		})
	}
	if g.GitSyncPasswordSecretPath != "" && g.GitSyncPasswordSecretName != "" {
		envs = append(envs, k8s.Env{
			Name: "GIT_SYNC_PASSWORD",
			ValueFrom: k8s.ValueFrom{
				SecretKeyRef: k8s.SecretKeyRef{
					Key:  g.GitSyncPasswordSecretPath,
					Name: g.GitSyncPasswordSecretName,
				},
			},
		})
	}
	return k8s.Container{
		Name:  k8s.ContainerName(fmt.Sprintf("git-sync-%s", g.MountName)),
		Image: k8s.Image(g.image().String()),
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
		Args: []k8s.Arg{
			"-logtostderr",
			"-v=4",
		},
		Env: envs,
		VolumeMounts: []k8s.ContainerMount{
			{
				Name: g.MountName,
				Path: "/target",
			},
		},
	}
}
