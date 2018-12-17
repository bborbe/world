// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"context"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type KafkaLatestVersions struct {
	Context      k8s.Context
	Domain       k8s.IngressHost
	Requirements []world.Configuration
	Replicas     k8s.Replicas
	StorageClass k8s.StorageClassName
	AccessMode   k8s.AccessMode
}

func (k *KafkaLatestVersions) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		k.Context,
		k.Domain,
		k.Replicas,
		k.StorageClass,
		k.AccessMode,
	)
}

func (k *KafkaLatestVersions) Applier() (world.Applier, error) {
	return nil, nil
}

func (k *KafkaLatestVersions) Children() []world.Configuration {
	var result []world.Configuration
	result = append(result, k.Requirements...)
	result = append(result, k.app()...)
	return result
}

func (k *KafkaLatestVersions) app() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/kafka-latest-versions",
		Tag:        "2.4.2",
	}
	port := deployer.Port{
		Port:     8080,
		Name:     "http",
		Protocol: "TCP",
	}
	return []world.Configuration{
		&k8s.NamespaceConfiguration{
			Context: k.Context,
			Namespace: k8s.Namespace{
				ApiVersion: "v1",
				Kind:       "Namespace",
				Metadata: k8s.Metadata{
					Namespace: "kafka-latest-versions",
					Name:      "kafka-latest-versions",
				},
			},
		},
		&build.KafkaLatestVersions{
			Image: image,
		},
		&k8s.StatefulSetConfiguration{
			Context: k.Context,
			StatefulSet: k8s.StatefulSet{
				ApiVersion: "apps/v1beta1",
				Kind:       "StatefulSet",
				Metadata: k8s.Metadata{
					Namespace: "kafka-latest-versions",
					Name:      "kafka-latest-versions",
					Labels: k8s.Labels{
						"app": "kafka-latest-versions",
					},
				},
				Spec: k8s.StatefulSetSpec{
					ServiceName: "kafka-latest-versions-headless",
					Replicas:    k.Replicas,
					Template: k8s.PodTemplate{
						Metadata: k8s.Metadata{
							Labels: k8s.Labels{
								"app": "kafka-latest-versions",
							},
							Annotations: k8s.Annotations{
								"prometheus.io/path":   "/metrics",
								"prometheus.io/port":   port.Port.String(),
								"prometheus.io/scheme": "http",
								"prometheus.io/scrape": "true",
							},
						},
						Spec: k8s.PodSpec{
							Containers: []k8s.Container{
								{
									Name:  "server",
									Image: k8s.Image(image.String()),
									Ports: []k8s.ContainerPort{port.ContainerPort()},
									Args:  []k8s.Arg{"-v=2"},
									Env: []k8s.Env{
										{
											Name:  "PORT",
											Value: port.Port.String(),
										},
										{
											Name:  "KAFKA_BROKERS",
											Value: "kafka-cp-kafka-headless.kafka.svc.cluster.local:9092",
										},
										{
											Name:  "KAFKA_LATEST_VERSION_TOPIC",
											Value: "application-version-latest",
										},
										{
											Name:  "KAFKA_AVAILABLE_VERSION_TOPIC",
											Value: "application-version-available",
										},
										{
											Name:  "KAFKA_SCHEMA_REGISTRY_URL",
											Value: "http://kafka-cp-schema-registry.kafka.svc.cluster.local:8081",
										},
										{
											Name:  "DATADIR",
											Value: "/data",
										},
									},
									VolumeMounts: []k8s.ContainerMount{
										{
											Path: "/data",
											Name: "datadir",
										},
									},
									Resources: k8s.Resources{
										Limits: k8s.ContainerResource{
											Cpu:    "500m",
											Memory: "50Mi",
										},
										Requests: k8s.ContainerResource{
											Cpu:    "25m",
											Memory: "25Mi",
										},
									},
									LivenessProbe: k8s.Probe{
										HttpGet: k8s.HttpGet{
											Path:   "/healthz",
											Port:   port.Port,
											Scheme: "HTTP",
										},
										InitialDelaySeconds: 10,
										SuccessThreshold:    1,
										FailureThreshold:    5,
										TimeoutSeconds:      5,
									},
									ReadinessProbe: k8s.Probe{
										HttpGet: k8s.HttpGet{
											Path:   "/readiness",
											Port:   port.Port,
											Scheme: "HTTP",
										},
										InitialDelaySeconds: 3,
										TimeoutSeconds:      5,
									},
								},
							},
						},
					},
					VolumeClaimTemplates: []k8s.VolumeClaimTemplate{
						{
							Metadata: k8s.Metadata{
								Name: "datadir",
								Annotations: map[string]string{
									"volume.alpha.kubernetes.io/storage-class": k.StorageClass.String(),
									"volume.beta.kubernetes.io/storage-class":  k.StorageClass.String(),
								},
							},
							Spec: k8s.VolumeClaimTemplatesSpec{
								AccessModes: []k8s.AccessMode{k.AccessMode},
								Resources: k8s.VolumeClaimTemplatesSpecResources{
									Requests: k8s.VolumeClaimTemplatesSpecResourcesRequests{
										Storage: "1Gi",
									},
								},
							},
						},
					},
					UpdateStrategy: k8s.UpdateStrategy{
						Type: "RollingUpdate",
					},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   k.Context,
			Namespace: "kafka-latest-versions",
			Name:      "kafka-latest-versions",
			Ports:     []deployer.Port{port},
		},
		&deployer.IngressDeployer{
			Context:   k.Context,
			Namespace: "kafka-latest-versions",
			Name:      "kafka-latest-versions",
			Port:      "http",
			Domains:   k8s.IngressHosts{k.Domain},
		},
	}
}
