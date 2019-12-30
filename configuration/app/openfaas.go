// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"context"

	"github.com/bborbe/world/pkg/dns"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/network"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type OpenFaas struct {
	Context k8s.Context
	Domain  k8s.IngressHost
	IP      network.IP
}

func (o *OpenFaas) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		o.Context,
		o.Domain,
		o.IP,
	)
}

func (o *OpenFaas) Applier() (world.Applier, error) {
	return nil, nil
}

func (o *OpenFaas) Children() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguraionBuilder().WithApplier(
			&dns.Server{
				Host:    "ns.rocketsource.de",
				KeyPath: "/Users/bborbe/.dns/home.benjamin-borbe.de.key",
				List: []dns.Entry{
					{
						Host: network.Host(o.Domain.String()),
						IP:   o.IP,
					},
				},
			},
		),
		&k8s.NamespaceConfiguration{
			Context: o.Context,
			Namespace: k8s.Namespace{
				ApiVersion: "v1",
				Kind:       "Namespace",
				Metadata: k8s.Metadata{
					Namespace: "openfaas",
					Name:      "openfaas",
				},
			},
		},
		&k8s.NamespaceConfiguration{
			Context: o.Context,
			Namespace: k8s.Namespace{
				ApiVersion: "v1",
				Kind:       "Namespace",
				Metadata: k8s.Metadata{
					Namespace: "openfaas-fn",
					Name:      "openfaas-fn",
				},
			},
		},
		&k8s.IngresseConfiguration{
			Context: o.Context,
			Ingress: k8s.Ingress{
				ApiVersion: "extensions/v1beta1",
				Kind:       "Ingress",
				Metadata: k8s.Metadata{
					Namespace: "openfaas",
					Name:      "openfaas",
					Annotations: k8s.Annotations{
						"kubernetes.io/ingress.class": "traefik",
						"traefik.frontend.priority":   "10000",
					}},
				Spec: k8s.IngressSpec{
					Rules: []k8s.IngressRule{
						{
							Host: o.Domain,
							Http: k8s.IngressHttp{
								Paths: []k8s.IngressPath{
									{
										Backends: k8s.IngressBackend{
											ServiceName: "gateway",
											ServicePort: "http",
										},
										Path: "/",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
