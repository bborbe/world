// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service

import (
	"context"
	"fmt"

	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type ClusterAdmin struct {
	Context k8s.Context
}

func (c *ClusterAdmin) Children() []world.Configuration {
	return []world.Configuration{
		&k8s.ClusterRoleBindingConfiguration{
			Context: c.Context,
			ClusterRoleBinding: k8s.ClusterRoleBinding{
				ApiVersion: "rbac.authorization.k8s.io/v1",
				Kind:       "ClusterRoleBinding",
				Metadata: k8s.Metadata{
					Name: "admin",
				},
				Subjects: []k8s.Subject{
					{
						Kind:     "User",
						Name:     fmt.Sprintf("%s-admin", c.Context),
						ApiGroup: "rbac.authorization.k8s.io",
					},
				},
				RoleRef: k8s.RoleRef{
					Kind:     "ClusterRole",
					Name:     "cluster-admin",
					ApiGroup: "rbac.authorization.k8s.io",
				},
			},
		},
	}
}
func (c *ClusterAdmin) Applier() (world.Applier, error) {
	return nil, nil
}

func (c *ClusterAdmin) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		c.Context,
	)
}
