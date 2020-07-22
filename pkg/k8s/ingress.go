// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package k8s

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

func BuildIngressConfigurationWithCertManager(
	context Context,
	namespace NamespaceName,
	name MetadataName,
	serviceName IngressBackendServiceName,
	servicePort IngressBackendServicePort,
	path IngressPathPath,
	hosts ...IngressHost,
) *IngresseConfiguration {
	result := &IngresseConfiguration{
		Context: context,
		Ingress: Ingress{
			ApiVersion: "extensions/v1beta1",
			Kind:       "Ingress",
			Metadata: Metadata{
				Namespace: namespace,
				Name:      name,
				Annotations: Annotations{
					"kubernetes.io/ingress.class":                    "nginx",
					"nginx.ingress.kubernetes.io/force-ssl-redirect": "true",
					"cert-manager.io/cluster-issuer":                 "letsencrypt-http-live",
				},
			},
		},
	}
	for _, host := range hosts {
		result.Ingress.Spec.TLS = append(result.Ingress.Spec.TLS, IngressTLS{
			Hosts: []string{
				host.String(),
			},
			SecretName: host.String(),
		})
		result.Ingress.Spec.Rules = append(result.Ingress.Spec.Rules, IngressRule{
			Host: host,
			Http: IngressHttp{
				Paths: []IngressPath{
					{
						Backends: IngressBackend{
							ServiceName: serviceName,
							ServicePort: servicePort,
						},
						Path: path,
					},
				},
			},
		})
	}
	return result
}

type IngresseConfiguration struct {
	Context      Context
	Ingress      Ingress
	Requirements []world.Configuration
}

func (i *IngresseConfiguration) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		i.Context,
		i.Ingress,
	)
}

func (i *IngresseConfiguration) Applier() (world.Applier, error) {
	return &IngressApplier{
		Context: i.Context,
		Ingress: i.Ingress,
	}, nil
}

func (i *IngresseConfiguration) Children() []world.Configuration {
	return i.Requirements
}

type IngressApplier struct {
	Context Context
	Ingress Ingress
}

func (i *IngressApplier) Satisfied(ctx context.Context) (bool, error) {
	return false, nil
}

func (i *IngressApplier) Apply(ctx context.Context) error {
	deployer := &Deployer{
		Context: i.Context,
		Data:    i.Ingress,
	}
	return deployer.Apply(ctx)
}

func (i *IngressApplier) Validate(ctx context.Context) error {
	if i.Context == "" {
		return errors.New("context missing")
	}
	return i.Ingress.Validate(ctx)
}

type Ingress struct {
	ApiVersion ApiVersion  `yaml:"apiVersion"`
	Kind       Kind        `yaml:"kind"`
	Metadata   Metadata    `yaml:"metadata"`
	Spec       IngressSpec `yaml:"spec"`
}

func (s Ingress) String() string {
	return fmt.Sprintf("%s/%s to %s", s.Kind, s.Metadata.Name, s.Metadata.Namespace)
}

func (s Ingress) Validate(ctx context.Context) error {
	if s.ApiVersion != "extensions/v1beta1" {
		return errors.New("invalid ApiVersion")
	}
	if s.Kind != "Ingress" {
		return errors.New("invalid Kind")
	}
	return nil
}

type IngressSpec struct {
	Rules []IngressRule `yaml:"rules"`
	TLS   []IngressTLS  `yaml:"tls,omitempty"`
}

func (i IngressSpec) Validate(ctx context.Context) error {
	return nil
}

type IngressTLS struct {
	Hosts      []string `yaml:"hosts,omitempty"`
	SecretName string   `yaml:"secretName,omitempty"`
}

func (i IngressTLS) Validate(ctx context.Context) error {
	return nil
}

type IngressHosts []IngressHost

func (i IngressHosts) Validate(ctx context.Context) error {
	if len(i) == 0 {
		return errors.New("IngressHosts empty")
	}
	for _, domain := range i {
		if err := domain.Validate(ctx); err != nil {
			return err
		}
	}
	return nil
}

type IngressHost string

func (i IngressHost) String() string {
	return string(i)
}

func (i IngressHost) Validate(ctx context.Context) error {
	if i == "" {
		return errors.New("ingressHost empty")
	}
	if strings.ContainsRune(i.String(), '_') {
		return errors.New("invalid char in ingressHost")
	}
	return nil
}

type IngressRule struct {
	Host IngressHost `yaml:"host"`
	Http IngressHttp `yaml:"http"`
}

type IngressHttp struct {
	Paths []IngressPath `yaml:"paths"`
}

type IngressPathPath string

type IngressPath struct {
	Backends IngressBackend  `yaml:"backend"`
	Path     IngressPathPath `yaml:"path"`
}

type IngressBackendServiceName string

type IngressBackendServicePort string

type IngressBackend struct {
	ServiceName IngressBackendServiceName `yaml:"serviceName"`
	ServicePort IngressBackendServicePort `yaml:"servicePort"`
}
