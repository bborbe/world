// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package deployer

import (
	"context"

	"github.com/pkg/errors"

	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type IngressDeployer struct {
	Context      k8s.Context
	Namespace    k8s.NamespaceName
	Name         k8s.MetadataName
	Domains      k8s.IngressHosts
	Port         k8s.PortName
	Requirements []world.Configuration
}

func (t *IngressDeployer) Validate(ctx context.Context) error {
	if len(t.Domains) == 0 {
		return errors.New("Domains empty")
	}
	return validation.Validate(
		ctx,
		t.Context,
		t.Namespace,
		t.Name,
		t.Port,
	)
}

func (i *IngressDeployer) Applier() (world.Applier, error) {
	return &k8s.IngressApplier{
		Context: i.Context,
		Ingress: i.ingress(),
	}, nil
}

func (i *IngressDeployer) Children(ctx context.Context) (world.Configurations, error) {
	return i.Requirements, nil
}

func (i *IngressDeployer) ingress() k8s.Ingress {
	ingress := k8s.Ingress{
		ApiVersion: "extensions/v1beta1",
		Kind:       "Ingress",
		Metadata: k8s.Metadata{
			Namespace: i.Namespace,
			Name:      i.Name,
			Labels: k8s.Labels{
				"app": i.Namespace.String(),
			},
		},
		Spec: k8s.IngressSpec{},
	}
	for _, domain := range i.Domains {
		ingress.Spec.Rules = append(ingress.Spec.Rules, k8s.IngressRule{
			Host: domain,
			Http: k8s.IngressHttp{
				Paths: []k8s.IngressPath{
					{
						Path: "/",
						Backends: k8s.IngressBackend{
							ServiceName: k8s.IngressBackendServiceName(i.Name),
							ServicePort: k8s.IngressBackendServicePort(i.Port),
						},
					},
				},
			},
		})
	}
	return ingress
}
