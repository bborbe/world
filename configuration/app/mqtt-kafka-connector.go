// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"context"
	"strings"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type MqttKafkaConnector struct {
	Context      k8s.Context
	MqttUser     deployer.SecretValue
	MqttPassword deployer.SecretValue
	MqttTopic    string
	KafkaTopic   string
	MqttBroker   string
	KafkaBrokers []string
}

func (m *MqttKafkaConnector) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		m.Context,
		m.MqttUser,
		m.MqttPassword,
	)
}

func (m *MqttKafkaConnector) Applier() (world.Applier, error) {
	return nil, nil
}

func (m *MqttKafkaConnector) Children() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/mqtt-kafka-connector",
		Tag:        "1.0.0",
	}
	port := deployer.Port{
		Port:     8080,
		Name:     "http",
		Protocol: "TCP",
	}
	return []world.Configuration{
		&k8s.NamespaceConfiguration{
			Context: m.Context,
			Namespace: k8s.Namespace{
				ApiVersion: "v1",
				Kind:       "Namespace",
				Metadata: k8s.Metadata{
					Namespace: "mqtt-kafka-connector",
					Name:      "mqtt-kafka-connector",
				},
			},
		},
		world.NewConfiguraionBuilder().WithApplier(
			&deployer.SecretApplier{
				Context:   m.Context,
				Namespace: "mqtt-kafka-connector",
				Name:      "mqtt",
				Secrets: deployer.Secrets{
					"user":     m.MqttUser,
					"password": m.MqttPassword,
				},
			},
		),
		&deployer.DeploymentDeployer{
			Context:   m.Context,
			Namespace: "mqtt-kafka-connector",
			Name:      "mqtt-kafka-connector",
			Strategy: k8s.DeploymentStrategy{
				Type: "RollingUpdate",
				RollingUpdate: k8s.DeploymentStrategyRollingUpdate{
					MaxSurge:       1,
					MaxUnavailable: 1,
				},
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name: "collector",
					Args: []k8s.Arg{"-v=2"},
					Env: []k8s.Env{
						{
							Name:  "PORT",
							Value: port.Port.String(),
						},
						{
							Name:  "KAFKA_BROKERS",
							Value: strings.Join(m.KafkaBrokers, ","),
						},
						{
							Name:  "KAFKA_TOPIC",
							Value: m.KafkaTopic,
						},
						{
							Name:  "MQTT_BROKER",
							Value: m.MqttBroker,
						},
						{
							Name: "MQTT_USER",
							ValueFrom: k8s.ValueFrom{
								SecretKeyRef: k8s.SecretKeyRef{
									Key:  "user",
									Name: "mqtt",
								},
							},
						},
						{
							Name: "MQTT_PASSWORD",
							ValueFrom: k8s.ValueFrom{
								SecretKeyRef: k8s.SecretKeyRef{
									Key:  "password",
									Name: "mqtt",
								},
							},
						},
						{
							Name:  "MQTT_TOPIC",
							Value: m.MqttTopic,
						},
					},
					Image: image,
					Requirement: &build.MqttKafkaConnector{
						Image: image,
					},
					Ports: []deployer.Port{port},
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "250m",
							Memory: "25Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "10m",
							Memory: "10Mi",
						},
					},
					LivenessProbe: k8s.Probe{
						HttpGet: k8s.HttpGet{
							Path:   "/",
							Port:   port.Port,
							Scheme: "HTTP",
						},
						InitialDelaySeconds: 60,
						SuccessThreshold:    1,
						FailureThreshold:    5,
						TimeoutSeconds:      5,
					},
					ReadinessProbe: k8s.Probe{
						HttpGet: k8s.HttpGet{
							Path:   "/",
							Port:   port.Port,
							Scheme: "HTTP",
						},
						InitialDelaySeconds: 3,
						TimeoutSeconds:      5,
					},
				},
			},
		},
	}
}
