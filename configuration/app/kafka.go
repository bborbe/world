// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Kafka struct {
	AccessMode        k8s.AccessMode
	Context           k8s.Context
	DisableConnect    bool
	DisableRest       bool
	KafkaReplicas     k8s.Replicas
	KafkaStorage      k8s.Storage
	StorageClass      k8s.StorageClassName
	ZookeeperReplicas k8s.Replicas
	ZookeeperStorage  k8s.Storage
	Version           docker.Tag
}

func (k *Kafka) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		k.Context,
		k.StorageClass,
		k.AccessMode,
		k.ZookeeperReplicas,
		k.KafkaReplicas,
		k.Version,
	)
}
func (k *Kafka) Children() []world.Configuration {
	result := []world.Configuration{
		&k8s.NamespaceConfiguration{
			Context: k.Context,
			Namespace: k8s.Namespace{
				ApiVersion: "v1",
				Kind:       "Namespace",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka",
				},
			},
		},
	}
	result = append(result, k.zookeeper()...)
	result = append(result, k.kafka()...)
	result = append(result, k.ksql()...)
	result = append(result, k.schemaRegistry()...)
	if !k.DisableConnect {
		result = append(result, k.connect()...)
	}
	if !k.DisableRest {
		result = append(result, k.rest()...)
	}
	return result
}

func (k *Kafka) connect() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/cp-kafka-connect",
		Tag:        k.Version,
	}
	return []world.Configuration{
		&build.CpKafkaConnect{
			Image: image,
		},
		&k8s.DeploymentConfiguration{
			Context: k.Context,
			Deployment: k8s.Deployment{
				ApiVersion: "apps/v1",
				Kind:       "Deployment",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka-cp-kafka-connect",
					Labels: k8s.Labels{
						"app": "cp-kafka-connect",
					},
				},
				Spec: k8s.DeploymentSpec{
					Replicas: 1,
					Selector: k8s.LabelSelector{
						MatchLabels: k8s.Labels{
							"app": "cp-kafka-connect",
						},
					},
					Template: k8s.PodTemplate{
						Metadata: k8s.Metadata{
							Labels: k8s.Labels{
								"app": "cp-kafka-connect",
							},
							Annotations: k8s.Annotations{
								"prometheus.io/scrape": "true",
								"prometheus.io/port":   "5556",
							},
						},
						Spec: k8s.PodSpec{
							Containers: []k8s.Container{
								{
									Name:  "prometheus-jmx-exporter",
									Image: "solsson/kafka-prometheus-jmx-exporter@sha256:a23062396cd5af1acdf76512632c20ea6be76885dfc20cd9ff40fb23846557e8",
									Command: []k8s.Command{
										"java",
										"-XX:+UnlockExperimentalVMOptions",
										"-XX:+UseCGroupMemoryLimitForHeap",
										"-XX:MaxRAMFraction=1",
										"-XshowSettings:vm",
										"-jar",
										"jmx_prometheus_httpserver.jar",
										"5556",
										"/etc/jmx-kafka-connect/jmx-kafka-connect-prometheus.yml",
									},
									Ports: []k8s.ContainerPort{
										{
											ContainerPort: 5556,
										},
									},
									VolumeMounts: []k8s.ContainerMount{
										{
											Path: "/etc/jmx-kafka-connect",
											Name: "jmx-config",
										},
									},
									Resources: k8s.Resources{
										Limits: k8s.ContainerResource{
											Cpu:    "2000m",
											Memory: "2000Mi",
										},
										Requests: k8s.ContainerResource{
											Cpu:    "10m",
											Memory: "10Mi",
										},
									},
								},
								{
									Name:  "cp-kafka-connect-server",
									Image: k8s.Image(image.String()),
									Env: []k8s.Env{
										{
											Name: "CONNECT_REST_ADVERTISED_HOST_NAME",
											ValueFrom: k8s.ValueFrom{
												FieldRef: k8s.FieldRef{
													FieldPath: "metadata.name",
												},
											},
										},
										{
											Name:  "CONNECT_CONFIG_STORAGE_REPLICATION_FACTOR",
											Value: k.KafkaReplicas.String(),
										},
										{
											Name:  "CONNECT_OFFSET_STORAGE_REPLICATION_FACTOR",
											Value: k.KafkaReplicas.String(),
										},
										{
											Name:  "CONNECT_STATUS_STORAGE_REPLICATION_FACTOR",
											Value: k.KafkaReplicas.String(),
										},
										{
											Name:  "CONNECT_PLUGIN_PATH",
											Value: "/usr/share/java",
										},
										{
											Name:  "CONNECT_BOOTSTRAP_SERVERS",
											Value: "PLAINTEXT://kafka-cp-kafka-headless:9092",
										},
										{
											Name:  "CONNECT_GROUP_ID",
											Value: "kafka",
										},
										{
											Name:  "CONNECT_CONFIG_STORAGE_TOPIC",
											Value: "kafka-cp-kafka-connect-config",
										},
										{
											Name:  "CONNECT_OFFSET_STORAGE_TOPIC",
											Value: "kafka-cp-kafka-connect-offset",
										},
										{
											Name:  "CONNECT_STATUS_STORAGE_TOPIC",
											Value: "kafka-cp-kafka-connect-status",
										},
										{
											Name:  "CONNECT_KEY_CONVERTER_SCHEMA_REGISTRY_URL",
											Value: "http://kafka-cp-schema-registry:8081",
										},
										{
											Name:  "CONNECT_VALUE_CONVERTER_SCHEMA_REGISTRY_URL",
											Value: "http://kafka-cp-schema-registry:8081",
										},
										{
											Name:  "CONNECT_KEY_CONVERTER",
											Value: "io.confluent.connect.avro.AvroConverter",
										},
										{
											Name:  "CONNECT_VALUE_CONVERTER",
											Value: "io.confluent.connect.avro.AvroConverter",
										},
										{
											Name:  "CONNECT_INTERNAL_KEY_CONVERTER",
											Value: "org.apache.kafka.connect.json.JsonConverter",
										},
										{
											Name:  "CONNECT_INTERNAL_VALUE_CONVERTER",
											Value: "org.apache.kafka.connect.json.JsonConverter",
										},
										{
											Name:  "KAFKA_JMX_PORT",
											Value: "5555",
										},
									},
									Ports: []k8s.ContainerPort{
										{
											ContainerPort: 8083,
											Name:          "kafka-connect",
											Protocol:      "TCP",
										},
										{
											ContainerPort: 5555,
											Name:          "jmx",
										},
									},
									Resources: k8s.Resources{
										Limits: k8s.ContainerResource{
											Cpu:    "2000m",
											Memory: "2000Mi",
										},
										Requests: k8s.ContainerResource{
											Cpu:    "10m",
											Memory: "10Mi",
										},
									},
									ImagePullPolicy: "IfNotPresent",
								},
							},
							Volumes: []k8s.PodVolume{
								{
									Name: "jmx-config",
									ConfigMap: k8s.PodVolumeConfigMap{
										Name: "kafka-cp-kafka-connect-jmx-configmap",
									},
								},
							},
						},
					},
				},
			},
		},
		&k8s.ConfigMapConfiguration{
			Context: k.Context,
			ConfigMap: k8s.ConfigMap{
				ApiVersion: "v1",
				Kind:       "ConfigMap",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka-cp-kafka-connect-jmx-configmap",
					Labels: k8s.Labels{
						"app": "cp-kafka-connect",
					},
				},
				Data: k8s.ConfigMapData{
					"jmx-kafka-connect-prometheus.yml": kafkaCpKafkaConnectJmxConfigmap,
				},
			},
		},
		&k8s.ServiceConfiguration{
			Context: k.Context,
			Service: k8s.Service{
				ApiVersion: "v1",
				Kind:       "Service",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka-cp-kafka-connect",
					Labels: k8s.Labels{
						"app": "cp-kafka-connect",
					},
				},
				Spec: k8s.ServiceSpec{
					Ports: []k8s.ServicePort{
						{
							Name: "kafka-connect",
							Port: 8083,
						},
					},
					Selector: k8s.ServiceSelector{
						"app": "cp-kafka-connect",
					},
				},
			},
		},
	}
}
func (k *Kafka) kafka() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/cp-kafka",
		Tag:        k.Version,
	}
	return []world.Configuration{
		&build.CpKafka{
			Image: image,
		},
		&k8s.ConfigMapConfiguration{
			Context: k.Context,
			ConfigMap: k8s.ConfigMap{
				ApiVersion: "v1",
				Kind:       "ConfigMap",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka-cp-kafka-jmx-configmap",
				},
				Data: k8s.ConfigMapData{
					"jmx-kafka-prometheus.yml": kafkaCpKafkaJmxConfigmap,
				},
			},
		},
		&k8s.StatefulSetConfiguration{
			Context: k.Context,
			StatefulSet: k8s.StatefulSet{
				ApiVersion: "apps/v1beta1",
				Kind:       "StatefulSet",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka-cp-kafka",
				},
				Spec: k8s.StatefulSetSpec{
					ServiceName: "kafka-cp-kafka-headless",
					Replicas:    k.KafkaReplicas,
					Template: k8s.PodTemplate{
						Metadata: k8s.Metadata{
							Labels: k8s.Labels{
								"app": "cp-kafka",
							},
							Annotations: k8s.Annotations{
								"prometheus.io/scrape": "true",
								"prometheus.io/port":   "5556",
							},
						},
						Spec: k8s.PodSpec{
							Affinity: k8s.Affinity{
								NodeAffinity: k8s.NodeAffinity{
									RequiredDuringSchedulingIgnoredDuringExecution: k8s.NodeSelector{
										NodeSelectorTerms: []k8s.NodeSelectorTerm{
											{
												MatchExpressions: []k8s.NodeSelectorRequirement{
													{
														Key:      "cloud.google.com/gke-preemptible",
														Operator: "DoesNotExist",
													},
												},
											},
										},
									},
								},
							},
							Containers: []k8s.Container{
								{
									Name:  "cp-kafka-broker",
									Image: k8s.Image(image.String()),
									Command: []k8s.Command{
										"sh",
										"-exc",
										"export KAFKA_BROKER_ID=${HOSTNAME##*-} && \\\nexport KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://${POD_IP}:9092,EXTERNAL://${HOST_IP}:$((31090 + ${KAFKA_BROKER_ID})) && \\\nexec /etc/confluent/docker/run\n",
									},
									Env: []k8s.Env{
										{
											Name: "POD_IP",
											ValueFrom: k8s.ValueFrom{
												FieldRef: k8s.FieldRef{
													FieldPath: "status.podIP",
												},
											},
										},
										{
											Name: "HOST_IP",
											ValueFrom: k8s.ValueFrom{
												FieldRef: k8s.FieldRef{
													FieldPath: "status.hostIP",
												},
											},
										},
										{
											Name:  "KAFKA_HEAP_OPTS",
											Value: "-Xms512M -Xmx512M",
										},
										{
											Name:  "KAFKA_ZOOKEEPER_CONNECT",
											Value: "kafka-cp-zookeeper-headless:2181",
										},
										{
											Name:  "KAFKA_ADVERTISED_LISTENERS",
											Value: "EXTERNAL://${HOST_IP}:$((31090 + ${KAFKA_BROKER_ID}))",
										},
										{
											Name:  "KAFKA_LISTENER_SECURITY_PROTOCOL_MAP",
											Value: "PLAINTEXT:PLAINTEXT,EXTERNAL:PLAINTEXT",
										},
										{
											Name:  "KAFKA_LOG_DIRS",
											Value: "/opt/kafka/data/logs",
										},
										{
											Name:  "KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR",
											Value: k.KafkaReplicas.String(),
										},
										{
											Name:  "KAFKA_JMX_PORT",
											Value: "5555",
										},
										{
											Name:  "KAFKA_AUTO_CREATE_TOPICS_ENABLE",
											Value: "false",
										},
									},
									Ports: []k8s.ContainerPort{
										{
											ContainerPort: 9092,
											Name:          "kafka",
										},
										{
											ContainerPort: 5555,
											Name:          "jmx",
										},
									},
									Resources: k8s.Resources{
										Limits: k8s.ContainerResource{
											Cpu:    "2000m",
											Memory: "2000Mi",
										},
										Requests: k8s.ContainerResource{
											Cpu:    "10m",
											Memory: "10Mi",
										},
									},
									VolumeMounts: []k8s.ContainerMount{
										{
											Path: "/opt/kafka/data",
											Name: "datadir",
										},
									},
									ImagePullPolicy: "IfNotPresent",
								},
								{
									Name:  "prometheus-jmx-exporter",
									Image: "solsson/kafka-prometheus-jmx-exporter@sha256:a23062396cd5af1acdf76512632c20ea6be76885dfc20cd9ff40fb23846557e8",
									Command: []k8s.Command{
										"java",
										"-XX:+UnlockExperimentalVMOptions",
										"-XX:+UseCGroupMemoryLimitForHeap",
										"-XX:MaxRAMFraction=1",
										"-XshowSettings:vm",
										"-jar",
										"jmx_prometheus_httpserver.jar",
										"5556",
										"/etc/jmx-kafka/jmx-kafka-prometheus.yml",
									},
									Ports: []k8s.ContainerPort{
										{
											ContainerPort: 5556,
										},
									},
									VolumeMounts: []k8s.ContainerMount{
										{
											Path: "/etc/jmx-kafka",
											Name: "jmx-config",
										},
									},
									Resources: k8s.Resources{
										Limits: k8s.ContainerResource{
											Cpu:    "2000m",
											Memory: "2000Mi",
										},
										Requests: k8s.ContainerResource{
											Cpu:    "10m",
											Memory: "10Mi",
										},
									},
								},
							},
							Volumes: []k8s.PodVolume{
								{
									Name: "jmx-config",
									ConfigMap: k8s.PodVolumeConfigMap{
										Name: "kafka-cp-kafka-jmx-configmap",
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
										Storage: k.KafkaStorage,
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
		&k8s.ServiceConfiguration{
			Context: k.Context,
			Service: k8s.Service{
				ApiVersion: "v1",
				Kind:       "Service",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka-cp-kafka-headless",
				},
				Spec: k8s.ServiceSpec{
					Ports: []k8s.ServicePort{
						{
							Name: "broker",
							Port: 9092,
						},
					},
					Selector: k8s.ServiceSelector{
						"app": "cp-kafka",
					},
					ClusterIP: "None",
				},
			},
		},
		&k8s.ServiceConfiguration{
			Context: k.Context,
			Service: k8s.Service{
				ApiVersion: "v1",
				Kind:       "Service",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka-cp-kafka",
				},
				Spec: k8s.ServiceSpec{
					Ports: []k8s.ServicePort{
						{
							Name: "broker",
							Port: 9092,
						},
					},
					Selector: k8s.ServiceSelector{
						"app": "cp-kafka",
					},
				},
			},
		},
	}
}
func (k *Kafka) rest() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/cp-kafka-rest",
		Tag:        k.Version,
	}
	return []world.Configuration{
		&build.CpKafkaRest{
			Image: image,
		},
		&k8s.DeploymentConfiguration{
			Context: k.Context,
			Deployment: k8s.Deployment{
				ApiVersion: "apps/v1",
				Kind:       "Deployment",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka-cp-kafka-rest",
					Labels: k8s.Labels{
						"app": "cp-kafka-rest",
					},
				},
				Spec: k8s.DeploymentSpec{
					Replicas:             1,
					RevisionHistoryLimit: 0,
					Selector: k8s.LabelSelector{
						MatchLabels: k8s.Labels{
							"app": "cp-kafka-rest",
						},
					},
					Template: k8s.PodTemplate{
						Metadata: k8s.Metadata{
							Labels: k8s.Labels{
								"app": "cp-kafka-rest",
							},
							Annotations: k8s.Annotations{
								"prometheus.io/scrape": "true",
								"prometheus.io/port":   "5556",
							},
						},
						Spec: k8s.PodSpec{
							Affinity: k8s.Affinity{
								NodeAffinity: k8s.NodeAffinity{
									RequiredDuringSchedulingIgnoredDuringExecution: k8s.NodeSelector{
										NodeSelectorTerms: []k8s.NodeSelectorTerm{
											{
												MatchExpressions: []k8s.NodeSelectorRequirement{
													{
														Key:      "cloud.google.com/gke-preemptible",
														Operator: "DoesNotExist",
													},
												},
											},
										},
									},
								},
							},
							Containers: []k8s.Container{
								{
									Name:  "prometheus-jmx-exporter",
									Image: "solsson/kafka-prometheus-jmx-exporter@sha256:a23062396cd5af1acdf76512632c20ea6be76885dfc20cd9ff40fb23846557e8",
									Command: []k8s.Command{
										"java",
										"-XX:+UnlockExperimentalVMOptions",
										"-XX:+UseCGroupMemoryLimitForHeap",
										"-XX:MaxRAMFraction=1",
										"-XshowSettings:vm",
										"-jar",
										"jmx_prometheus_httpserver.jar",
										"5556",
										"/etc/jmx-kafka-rest/jmx-kafka-rest-prometheus.yml",
									},
									Ports: []k8s.ContainerPort{
										{
											ContainerPort: 5556,
										},
									},
									Resources: k8s.Resources{
										Limits: k8s.ContainerResource{
											Cpu:    "2000m",
											Memory: "2000Mi",
										},
										Requests: k8s.ContainerResource{
											Cpu:    "10m",
											Memory: "10Mi",
										},
									},
									VolumeMounts: []k8s.ContainerMount{
										{
											Path: "/etc/jmx-kafka-rest",
											Name: "jmx-config",
										},
									},
								},
								{
									Name:  "cp-kafka-rest-server",
									Image: k8s.Image(image.String()),
									Env: []k8s.Env{
										{
											Name: "KAFKA_REST_HOST_NAME",
											ValueFrom: k8s.ValueFrom{
												SecretKeyRef: k8s.SecretKeyRef{},
												FieldRef: k8s.FieldRef{
													FieldPath: "metadata.name",
												},
											},
										},
										{
											Name:  "KAFKA_REST_ZOOKEEPER_CONNECT",
											Value: "kafka-cp-zookeeper-headless:2181",
										},
										{
											Name:  "KAFKA_REST_SCHEMA_REGISTRY_URL",
											Value: "http://kafka-cp-schema-registry:8081",
										},
										{
											Name:  "KAFKA_REST_JMX_PORT",
											Value: "5555",
										},
									},
									Ports: []k8s.ContainerPort{
										{
											ContainerPort: 8082,
											Name:          "rest-proxy",
											Protocol:      "TCP",
										},
										{
											ContainerPort: 5555,
											Name:          "jmx",
										},
									},
									ImagePullPolicy: "IfNotPresent",
									Resources: k8s.Resources{
										Limits: k8s.ContainerResource{
											Cpu:    "2000m",
											Memory: "2000Mi",
										},
										Requests: k8s.ContainerResource{
											Cpu:    "10m",
											Memory: "10Mi",
										},
									},
								},
							},
							Volumes: []k8s.PodVolume{
								{
									Name: "jmx-config",
									ConfigMap: k8s.PodVolumeConfigMap{
										Name: "kafka-cp-kafka-rest-jmx-configmap",
									},
								},
							},
						},
					},
				},
			},
		},
		&k8s.ConfigMapConfiguration{
			Context: k.Context,
			ConfigMap: k8s.ConfigMap{
				ApiVersion: "v1",
				Kind:       "ConfigMap",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka-cp-kafka-rest-jmx-configmap",
					Labels: k8s.Labels{
						"app": "cp-kafka-rest",
					},
				},
				Data: k8s.ConfigMapData{
					"jmx-kafka-rest-prometheus.yml": kafkaCpKafkaRestJmxConfigmap,
				},
			},
		},
		&k8s.ServiceConfiguration{
			Context: k.Context,
			Service: k8s.Service{
				ApiVersion: "v1",
				Kind:       "Service",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka-cp-kafka-rest",
					Labels: k8s.Labels{
						"app": "cp-kafka-rest",
					},
				},
				Spec: k8s.ServiceSpec{
					Ports: []k8s.ServicePort{
						{
							Name: "rest-proxy",
							Port: 8082,
						},
					},
					Selector: k8s.ServiceSelector{
						"app": "cp-kafka-rest",
					},
				},
			},
		},
	}
}
func (k *Kafka) ksql() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/cp-ksql-server",
		Tag:        k.Version,
	}
	return []world.Configuration{
		&build.CpKafkaKsql{
			Image: image,
		},
		&k8s.DeploymentConfiguration{
			Context: k.Context,
			Deployment: k8s.Deployment{
				ApiVersion: "apps/v1",
				Kind:       "Deployment",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka-cp-ksql-server",
					Labels: k8s.Labels{
						"app": "cp-ksql-server",
					},
				},
				Spec: k8s.DeploymentSpec{
					Replicas:             1,
					RevisionHistoryLimit: 0,
					Selector: k8s.LabelSelector{
						MatchLabels: k8s.Labels{
							"app": "cp-ksql-server",
						},
					},
					Strategy: k8s.DeploymentStrategy{
						RollingUpdate: k8s.DeploymentStrategyRollingUpdate{
							MaxSurge:       0,
							MaxUnavailable: 0},
					},
					Template: k8s.PodTemplate{
						Metadata: k8s.Metadata{
							Labels: k8s.Labels{
								"app": "cp-ksql-server",
							},
							Annotations: k8s.Annotations{
								"prometheus.io/scrape": "true",
								"prometheus.io/port":   "5556",
							},
						},
						Spec: k8s.PodSpec{
							Affinity: k8s.Affinity{
								NodeAffinity: k8s.NodeAffinity{
									RequiredDuringSchedulingIgnoredDuringExecution: k8s.NodeSelector{
										NodeSelectorTerms: []k8s.NodeSelectorTerm{
											{
												MatchExpressions: []k8s.NodeSelectorRequirement{
													{
														Key:      "cloud.google.com/gke-preemptible",
														Operator: "DoesNotExist",
													},
												},
											},
										},
									},
								},
							},
							Containers: []k8s.Container{
								{
									Name:  "cp-ksql-server",
									Image: k8s.Image(image.String()),
									Env: []k8s.Env{
										{
											Name:  "KSQL_BOOTSTRAP_SERVERS",
											Value: "PLAINTEXT://kafka-cp-kafka-headless:9092",
										},
										{
											Name:  "KSQL_KSQL_SERVICE_ID",
											Value: "kafka",
										},
										{
											Name:  "KSQL_LISTENERS",
											Value: "http://0.0.0.0:8088",
										},
										{
											Name:  "KSQL_JMX_PORT",
											Value: "5555",
										},
										{
											Name:  "KSQL_KSQL_SCHEMA_REGISTRY_URL",
											Value: "http://kafka-cp-schema-registry:8081",
										},
									},
									Ports: []k8s.ContainerPort{
										{
											ContainerPort: 8088,
											Name:          "server",
											Protocol:      "TCP",
										},
										{
											ContainerPort: 5555,
											Name:          "jmx",
										},
									},
									Resources: k8s.Resources{
										Limits: k8s.ContainerResource{
											Cpu:    "2000m",
											Memory: "2000Mi",
										},
										Requests: k8s.ContainerResource{
											Cpu:    "10m",
											Memory: "10Mi",
										},
									},
									ImagePullPolicy: "IfNotPresent",
								},
								{
									Name:  "prometheus-jmx-exporter",
									Image: "solsson/kafka-prometheus-jmx-exporter@sha256:a23062396cd5af1acdf76512632c20ea6be76885dfc20cd9ff40fb23846557e8",
									Command: []k8s.Command{
										"java",
										"-XX:+UnlockExperimentalVMOptions",
										"-XX:+UseCGroupMemoryLimitForHeap",
										"-XX:MaxRAMFraction=1",
										"-XshowSettings:vm",
										"-jar",
										"jmx_prometheus_httpserver.jar",
										"5556",
										"/etc/jmx-ksql-server/jmx-ksql-server-prometheus.yml",
									},
									Ports: []k8s.ContainerPort{
										{
											ContainerPort: 5556,
											HostPort:      0,
										},
									},
									Resources: k8s.Resources{
										Limits: k8s.ContainerResource{
											Cpu:    "2000m",
											Memory: "2000Mi",
										},
										Requests: k8s.ContainerResource{
											Cpu:    "10m",
											Memory: "10Mi",
										},
									},
									VolumeMounts: []k8s.ContainerMount{
										{
											Path:     "/etc/jmx-ksql-server",
											Name:     "jmx-config",
											ReadOnly: false},
									},
								},
							},
							Volumes: []k8s.PodVolume{
								{
									Name: "jmx-config",
									ConfigMap: k8s.PodVolumeConfigMap{
										Name: "kafka-cp-ksql-server-jmx-configmap",
									},
								},
							},
						},
					},
				},
			},
		},
		&k8s.ConfigMapConfiguration{
			Context: k.Context,
			ConfigMap: k8s.ConfigMap{
				ApiVersion: "v1",
				Kind:       "ConfigMap",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka-cp-ksql-server-jmx-configmap",
					Labels: k8s.Labels{
						"app": "cp-ksql-server",
					},
				},
				Data: k8s.ConfigMapData{
					"jmx-ksql-server-prometheus.yml": kafkaCpKsqlServerJmxConfigmap,
				},
			},
		},
		&k8s.ConfigMapConfiguration{
			Context: k.Context,
			ConfigMap: k8s.ConfigMap{
				ApiVersion: "v1",
				Kind:       "ConfigMap",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka-cp-ksql-server-ksql-queries-configmap",
					Labels: k8s.Labels{
						"app": "cp-ksql-server",
					},
				},
				Data: k8s.ConfigMapData{
					"queries.sql": ksqlQueries,
				},
			},
		},
		&k8s.ServiceConfiguration{
			Context: k.Context,
			Service: k8s.Service{
				ApiVersion: "v1",
				Kind:       "Service",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka-cp-ksql-server",
					Labels: k8s.Labels{
						"app": "cp-ksql-server",
					},
				},
				Spec: k8s.ServiceSpec{
					Ports: []k8s.ServicePort{
						{
							Name: "ksql-server",
							Port: 8088,
						},
					},
					Selector: k8s.ServiceSelector{
						"app": "cp-ksql-server",
					},
				},
			},
		},
	}
}
func (k *Kafka) schemaRegistry() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/cp-schema-registry",
		Tag:        k.Version,
	}
	return []world.Configuration{
		&build.CpKafkaSchemaRegistry{
			Image: image,
		},
		&k8s.DeploymentConfiguration{
			Context: k.Context,
			Deployment: k8s.Deployment{
				ApiVersion: "apps/v1",
				Kind:       "Deployment",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka-cp-schema-registry",
					Labels: k8s.Labels{
						"app": "cp-schema-registry",
					},
				},
				Spec: k8s.DeploymentSpec{
					Replicas:             1,
					RevisionHistoryLimit: 0,
					Selector: k8s.LabelSelector{
						MatchLabels: k8s.Labels{
							"app": "cp-schema-registry",
						},
					},
					Strategy: k8s.DeploymentStrategy{
						RollingUpdate: k8s.DeploymentStrategyRollingUpdate{
							MaxSurge:       0,
							MaxUnavailable: 0},
					},
					Template: k8s.PodTemplate{
						Metadata: k8s.Metadata{
							Labels: k8s.Labels{
								"app": "cp-schema-registry",
							},
							Annotations: k8s.Annotations{
								"prometheus.io/scrape": "true",
								"prometheus.io/port":   "5556",
							},
						},
						Spec: k8s.PodSpec{
							Affinity: k8s.Affinity{
								NodeAffinity: k8s.NodeAffinity{
									RequiredDuringSchedulingIgnoredDuringExecution: k8s.NodeSelector{
										NodeSelectorTerms: []k8s.NodeSelectorTerm{
											{
												MatchExpressions: []k8s.NodeSelectorRequirement{
													{
														Key:      "cloud.google.com/gke-preemptible",
														Operator: "DoesNotExist",
													},
												},
											},
										},
									},
								},
							},
							Containers: []k8s.Container{
								{
									Name:  "prometheus-jmx-exporter",
									Image: "solsson/kafka-prometheus-jmx-exporter@sha256:a23062396cd5af1acdf76512632c20ea6be76885dfc20cd9ff40fb23846557e8",
									Command: []k8s.Command{
										"java",
										"-XX:+UnlockExperimentalVMOptions",
										"-XX:+UseCGroupMemoryLimitForHeap",
										"-XX:MaxRAMFraction=1",
										"-XshowSettings:vm",
										"-jar",
										"jmx_prometheus_httpserver.jar",
										"5556",
										"/etc/jmx-schema-registry/jmx-schema-registry-prometheus.yml",
									},
									Ports: []k8s.ContainerPort{
										{
											ContainerPort: 5556,
										},
									},
									VolumeMounts: []k8s.ContainerMount{
										{
											Path: "/etc/jmx-schema-registry",
											Name: "jmx-config",
										},
									},
									Resources: k8s.Resources{
										Limits: k8s.ContainerResource{
											Cpu:    "2000m",
											Memory: "2000Mi",
										},
										Requests: k8s.ContainerResource{
											Cpu:    "10m",
											Memory: "10Mi",
										},
									},
								},
								{
									Name:  "cp-schema-registry-server",
									Image: k8s.Image(image.String()),
									Env: []k8s.Env{
										{
											Name: "SCHEMA_REGISTRY_HOST_NAME",
											ValueFrom: k8s.ValueFrom{
												FieldRef: k8s.FieldRef{
													FieldPath: "metadata.name",
												},
											},
										},
										{
											Name:  "SCHEMA_REGISTRY_KAFKASTORE_BOOTSTRAP_SERVERS",
											Value: "PLAINTEXT://kafka-cp-kafka-headless:9092",
										},
										{
											Name:  "SCHEMA_REGISTRY_KAFKASTORE_GROUP_ID",
											Value: "kafka",
										},
										{
											Name:  "SCHEMA_REGISTRY_MASTER_ELIGIBILITY",
											Value: "true",
										},
										{
											Name:  "JMX_PORT",
											Value: "5555",
										},
									},
									Ports: []k8s.ContainerPort{
										{
											ContainerPort: 8081,
											Name:          "schema-registry",
											Protocol:      "TCP",
										},
										{
											ContainerPort: 5555,
											Name:          "jmx",
										},
									},
									Resources: k8s.Resources{
										Limits: k8s.ContainerResource{
											Cpu:    "2000m",
											Memory: "2000Mi",
										},
										Requests: k8s.ContainerResource{
											Cpu:    "10m",
											Memory: "10Mi",
										},
									},
									ImagePullPolicy: "IfNotPresent",
								},
							},
							Volumes: []k8s.PodVolume{
								{
									Name: "jmx-config",
									ConfigMap: k8s.PodVolumeConfigMap{
										Name: "kafka-cp-schema-registry-jmx-configmap",
									},
								},
							},
						},
					},
				},
			},
		},
		&k8s.ConfigMapConfiguration{
			Context: k.Context,
			ConfigMap: k8s.ConfigMap{
				ApiVersion: "v1",
				Kind:       "ConfigMap",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka-cp-schema-registry-jmx-configmap",
					Labels: k8s.Labels{
						"app": "cp-schema-registry",
					},
				},
				Data: k8s.ConfigMapData{
					"jmx-schema-registry-prometheus.yml": kafkaCpSchemaRegistryJmxConfigmap,
				},
			},
		},
		&k8s.ServiceConfiguration{
			Context: k.Context,
			Service: k8s.Service{
				ApiVersion: "v1",
				Kind:       "Service",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka-cp-schema-registry",
					Labels: k8s.Labels{
						"app": "cp-schema-registry",
					},
				},
				Spec: k8s.ServiceSpec{
					Ports: []k8s.ServicePort{
						{
							Name: "schema-registry",
							Port: 8081,
						},
					},
					Selector: k8s.ServiceSelector{
						"app": "cp-schema-registry",
					},
				},
			},
		},
	}
}
func (k *Kafka) zookeeper() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/cp-zookeeper",
		Tag:        k.Version,
	}
	var zookeeperServerLists []string
	for i := k8s.Replicas(0); i < k.ZookeeperReplicas; i++ {
		addr := fmt.Sprintf("kafka-cp-zookeeper-%d.kafka-cp-zookeeper-headless.default.svc.cluster.local:2888:3888", i)
		zookeeperServerLists = append(zookeeperServerLists, addr)
	}
	return []world.Configuration{
		&build.CpZookeeper{
			Image: image,
		},
		&k8s.ServiceConfiguration{
			Context: k.Context,
			Service: k8s.Service{
				ApiVersion: "v1",
				Kind:       "Service",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka-cp-zookeeper-headless",
					Labels: k8s.Labels{
						"app": "cp-zookeeper",
					},
				},
				Spec: k8s.ServiceSpec{
					Ports: []k8s.ServicePort{
						{
							Name: "server",
							Port: 2888,
						},
						{
							Name: "leader-election",
							Port: 3888,
						},
					},
					Selector: k8s.ServiceSelector{
						"app": "cp-zookeeper",
					},
					ClusterIP: "None",
				},
			},
		},
		&k8s.ConfigMapConfiguration{
			Context: k.Context,
			ConfigMap: k8s.ConfigMap{
				ApiVersion: "v1",
				Kind:       "ConfigMap",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka-cp-zookeeper-jmx-configmap",
					Labels: k8s.Labels{
						"app": "cp-zookeeper",
					},
				},
				Data: k8s.ConfigMapData{
					"jmx-zookeeper-prometheus.yml": kafkaCpZookeeperJmxConfigmap,
				},
			},
		},
		&k8s.PodDisruptionBudgetConfiguration{
			Context: k.Context,
			PodDisruptionBudget: k8s.PodDisruptionBudget{
				ApiVersion: "policy/v1beta1",
				Kind:       "PodDisruptionBudget",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka-cp-zookeeper-pdb",
					Labels: k8s.Labels{
						"app": "cp-zookeeper",
					},
				},
				Spec: k8s.PodDisruptionBudgetSpec{
					MaxUnavailable: 1,
					MinAvailable:   0,
					Selector: k8s.LabelSelector{
						MatchLabels: k8s.Labels{
							"app": "cp-zookeeper",
						},
					},
				},
			},
		},
		&k8s.StatefulSetConfiguration{
			Context: k.Context,
			StatefulSet: k8s.StatefulSet{
				ApiVersion: "apps/v1beta1",
				Kind:       "StatefulSet",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka-cp-zookeeper",
					Labels: k8s.Labels{
						"app": "cp-zookeeper",
					},
				},
				Spec: k8s.StatefulSetSpec{
					ServiceName: "kafka-cp-zookeeper-headless",
					Replicas:    k8s.Replicas(k.ZookeeperReplicas),
					Template: k8s.PodTemplate{
						Metadata: k8s.Metadata{
							Labels: k8s.Labels{
								"app": "cp-zookeeper",
							},
							Annotations: k8s.Annotations{
								"prometheus.io/scrape": "true",
								"prometheus.io/port":   "5556",
							},
						},
						Spec: k8s.PodSpec{
							Affinity: k8s.Affinity{
								NodeAffinity: k8s.NodeAffinity{
									RequiredDuringSchedulingIgnoredDuringExecution: k8s.NodeSelector{
										NodeSelectorTerms: []k8s.NodeSelectorTerm{
											{
												MatchExpressions: []k8s.NodeSelectorRequirement{
													{
														Key:      "cloud.google.com/gke-preemptible",
														Operator: "DoesNotExist",
													},
												},
											},
										},
									},
								},
							},
							Containers: []k8s.Container{
								{
									Name:  "cp-zookeeper-server",
									Image: k8s.Image(image.String()),
									Command: []k8s.Command{
										"bash",
										"-c",
										"ZOOKEEPER_SERVER_ID=$((${HOSTNAME##*-}+1)) && /etc/confluent/docker/run",
									},
									Env: []k8s.Env{
										{
											Name:  "KAFKA_HEAP_OPTS",
											Value: "-Xms512M -Xmx512M",
										},
										{
											Name:  "KAFKA_JMX_PORT",
											Value: "5555",
										},
										{
											Name:  "ZOOKEEPER_TICK_TIME",
											Value: "2000",
										},
										{
											Name:  "ZOOKEEPER_SYNC_LIMIT",
											Value: "5",
										},
										{
											Name:  "ZOOKEEPER_INIT_LIMIT",
											Value: "10",
										},
										{
											Name:  "ZOOKEEPER_MAX_CLIENT_CNXNS",
											Value: "60",
										},
										{
											Name:  "ZOOKEEPER_AUTOPURGE_SNAP_RETAIN_COUNT",
											Value: "3",
										},
										{
											Name:  "ZOOKEEPER_AUTOPURGE_PURGE_INTERVAL",
											Value: "24",
										},
										{
											Name:  "ZOOKEEPER_CLIENT_PORT",
											Value: "2181",
										},
										{
											Name:  "ZOOKEEPER_SERVERS",
											Value: strings.Join(zookeeperServerLists, ";"),
										},
										{
											Name: "ZOOKEEPER_SERVER_ID",
											ValueFrom: k8s.ValueFrom{
												FieldRef: k8s.FieldRef{
													FieldPath: "metadata.name",
												},
											},
										},
									},
									Ports: []k8s.ContainerPort{
										{
											ContainerPort: 2181,
											Name:          "client",
										},
										{
											ContainerPort: 2888,
											Name:          "server",
										},
										{
											ContainerPort: 3888,
											Name:          "leader-election",
										},
										{
											ContainerPort: 5555,
											Name:          "jmx",
										},
									},
									VolumeMounts: []k8s.ContainerMount{
										{
											Path: "/var/lib/zookeeper/data",
											Name: "datadir",
										},
										{
											Path: "/var/lib/zookeeper/log",
											Name: "datalogdir",
										},
									},
									LivenessProbe: k8s.Probe{
										Exec: k8s.Exec{
											Command: []k8s.Command{
												"/bin/bash",
												"-c",
												"echo \"ruok\" | nc -w 2 -q 2 localhost 2181 | grep imok",
											},
										},
										InitialDelaySeconds: 1,
										TimeoutSeconds:      3,
									},
									Resources: k8s.Resources{
										Limits: k8s.ContainerResource{
											Cpu:    "2000m",
											Memory: "2000Mi",
										},
										Requests: k8s.ContainerResource{
											Cpu:    "10m",
											Memory: "10Mi",
										},
									},
									ImagePullPolicy: "IfNotPresent",
								},
								{
									Name:  "prometheus-jmx-exporter",
									Image: "solsson/kafka-prometheus-jmx-exporter@sha256:a23062396cd5af1acdf76512632c20ea6be76885dfc20cd9ff40fb23846557e8",
									Command: []k8s.Command{
										"java",
										"-XX:+UnlockExperimentalVMOptions",
										"-XX:+UseCGroupMemoryLimitForHeap",
										"-XX:MaxRAMFraction=1",
										"-XshowSettings:vm",
										"-jar",
										"jmx_prometheus_httpserver.jar",
										"5556",
										"/etc/jmx-zookeeper/jmx-zookeeper-prometheus.yml",
									},
									Ports: []k8s.ContainerPort{
										{
											ContainerPort: 5556,
										},
									},
									VolumeMounts: []k8s.ContainerMount{
										{
											Path: "/etc/jmx-zookeeper",
											Name: "jmx-config",
										},
									},
									Resources: k8s.Resources{
										Limits: k8s.ContainerResource{
											Cpu:    "2000m",
											Memory: "2000Mi",
										},
										Requests: k8s.ContainerResource{
											Cpu:    "10m",
											Memory: "10Mi",
										},
									},
								},
							},
							Volumes: []k8s.PodVolume{
								{
									Name: "jmx-config",
									ConfigMap: k8s.PodVolumeConfigMap{
										Name: "kafka-cp-zookeeper-jmx-configmap",
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
										Storage: k.ZookeeperStorage,
									},
								},
							},
						},
						{
							Metadata: k8s.Metadata{
								Name: "datalogdir",
								Annotations: map[string]string{
									"volume.alpha.kubernetes.io/storage-class": k.StorageClass.String(),
									"volume.beta.kubernetes.io/storage-class":  k.StorageClass.String(),
								},
							},
							Spec: k8s.VolumeClaimTemplatesSpec{
								AccessModes: []k8s.AccessMode{k.AccessMode},
								Resources: k8s.VolumeClaimTemplatesSpecResources{
									Requests: k8s.VolumeClaimTemplatesSpecResourcesRequests{
										Storage: k.ZookeeperStorage,
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
		&k8s.ServiceConfiguration{
			Context: k.Context,
			Service: k8s.Service{
				ApiVersion: "v1",
				Kind:       "Service",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka-cp-zookeeper",
					Labels: k8s.Labels{
						"app": "cp-zookeeper",
					},
				},
				Spec: k8s.ServiceSpec{
					Ports: []k8s.ServicePort{
						{
							Name: "client",
							Port: 2181,
						},
					},
					Selector: k8s.ServiceSelector{
						"app": "cp-zookeeper",
					},
				},
			},
		},
	}
}

func (k *Kafka) Applier() (world.Applier, error) {
	return nil, nil
}

const kafkaCpKafkaJmxConfigmap = `
jmxUrl: service:jmx:rmi:///jndi/rmi://localhost:5555/jmxrmi
lowercaseOutputName: true
lowercaseOutputLabelNames: true
ssl: false
rules:
- pattern : kafka.server<type=ReplicaManager, name=(.+)><>(Value|OneMinuteRate)
  name: "cp_kafka_server_replicamanager_$1"
- pattern : kafka.controller<type=KafkaController, name=(.+)><>Value
  name: "cp_kafka_controller_kafkacontroller_$1"
- pattern : kafka.server<type=BrokerTopicMetrics, name=(.+)><>OneMinuteRate
  name: "cp_kafka_server_brokertopicmetrics_$1"
- pattern : kafka.network<type=RequestMetrics, name=RequestsPerSec, request=(.+)><>OneMinuteRate
  name: "cp_kafka_network_requestmetrics_requestspersec_$1"
- pattern : kafka.network<type=SocketServer, name=NetworkProcessorAvgIdlePercent><>Value
  name: "cp_kafka_network_socketserver_networkprocessoravgidlepercent"
- pattern : kafka.server<type=ReplicaFetcherManager, name=MaxLag, clientId=(.+)><>Value
  name: "cp_kafka_server_replicafetchermanager_maxlag_$1"
- pattern : kafka.server<type=KafkaRequestHandlerPool, name=RequestHandlerAvgIdlePercent><>OneMinuteRate
  name: "cp_kafka_kafkarequesthandlerpool_requesthandleravgidlepercent"
- pattern : kafka.controller<type=ControllerStats, name=(.+)><>OneMinuteRate
  name: "cp_kafka_controller_controllerstats_$1"
- pattern : kafka.server<type=SessionExpireListener, name=(.+)><>OneMinuteRate
  name: "cp_kafka_server_sessionexpirelistener_$1"
`

const ksqlQueries = `
-- From http://docs.confluent.io/current/ksql/docs/tutorials/basics-docker.html#create-a-stream-and-table
    
-- Create a stream pageviews_original from the Kafka topic pageviews, specifying the value_format of DELIMITED
CREATE STREAM pageviews_original (viewtime bigint, userid varchar, pageid varchar) WITH (kafka_topic='pageviews', value_format='DELIMITED');

-- Create a table users_original from the Kafka topic users, specifying the value_format of JSON
CREATE TABLE users_original (registertime BIGINT, gender VARCHAR, regionid VARCHAR, userid VARCHAR) WITH (kafka_topic='users', value_format='JSON', key = 'userid');

-- Create a persistent query by using the CREATE STREAM keywords to precede the SELECT statement
CREATE STREAM pageviews_enriched AS SELECT users_original.userid AS userid, pageid, regionid, gender FROM pageviews_original LEFT JOIN users_original ON pageviews_original.userid = users_original.userid;

-- Create a new persistent query where a condition limits the streams content, using WHERE
CREATE STREAM pageviews_female AS SELECT * FROM pageviews_enriched WHERE gender = 'FEMALE';

-- Create a new persistent query where another condition is met, using LIKE
CREATE STREAM pageviews_female_like_89 WITH (kafka_topic='pageviews_enriched_r8_r9') AS SELECT * FROM pageviews_female WHERE regionid LIKE '%_8' OR regionid LIKE '%_9';

-- Create a new persistent query that counts the pageviews for each region and gender combination in a tumbling window of 30 seconds when the count is greater than one
CREATE TABLE pageviews_regions WITH (VALUE_FORMAT='avro') AS SELECT gender, regionid , COUNT(*) AS numusers FROM pageviews_enriched WINDOW TUMBLING (size 30 second) GROUP BY gender, regionid HAVING COUNT(*) > 1;
`

const kafkaCpZookeeperJmxConfigmap = `
jmxUrl: service:jmx:rmi:///jndi/rmi://localhost:5555/jmxrmi
lowercaseOutputName: true
lowercaseOutputLabelNames: true
ssl: false
rules:
- pattern: "org.apache.ZooKeeperService<name0=ReplicatedServer_id(\\d+)><>(\\w+)"
  name: "cp_zookeeper_$2"
- pattern: "org.apache.ZooKeeperService<name0=ReplicatedServer_id(\\d+), name1=replica.(\\d+)><>(\\w+)"
  name: "cp_zookeeper_$3"
  labels:
    replicaId: "$2"
- pattern: "org.apache.ZooKeeperService<name0=ReplicatedServer_id(\\d+), name1=replica.(\\d+), name2=(\\w+)><>(\\w+)"
  name: "cp_zookeeper_$4"
  labels:
    replicaId: "$2"
    memberType: "$3"
- pattern: "org.apache.ZooKeeperService<name0=ReplicatedServer_id(\\d+), name1=replica.(\\d+), name2=(\\w+), name3=(\\w+)><>(\\w+)"
  name: "cp_zookeeper_$4_$5"
  labels:
    replicaId: "$2"
    memberType: "$3"
`

const kafkaCpSchemaRegistryJmxConfigmap = `
jmxUrl: service:jmx:rmi:///jndi/rmi://localhost:5555/jmxrmi
lowercaseOutputName: true
lowercaseOutputLabelNames: true
ssl: false
rules:
- pattern : 'kafka.schema.registry<type=jetty-metrics>([^:]+):'
  name: "cp_kafka_schema_registry_jetty_metrics_$1"
- pattern : 'kafka.schema.registry<type=master-slave-role>([^:]+):'
  name: "cp_kafka_schema_registry_master_slave_role"
- pattern : 'kafka.schema.registry<type=jersey-metrics>([^:]+):'
  name: "cp_kafka_schema_registry_jersey_metrics_$1"
`

const kafkaCpKafkaConnectJmxConfigmap = `
jmxUrl: service:jmx:rmi:///jndi/rmi://localhost:5555/jmxrmi
lowercaseOutputName: true
lowercaseOutputLabelNames: true
ssl: false
rules:
- pattern : "kafka.connect<type=connect-worker-metrics>([^:]+):"
  name: "cp_kafka_connect_connect_worker_metrics_$1"
- pattern : "kafka.connect<type=connect-metrics, client-id=([^:]+)><>([^:]+)"
  name: "cp_kafka_connect_connect_metrics_$1_$2"
`

const kafkaCpKafkaRestJmxConfigmap = `
jmxUrl: service:jmx:rmi:///jndi/rmi://localhost:5555/jmxrmi
lowercaseOutputName: true
lowercaseOutputLabelNames: true
ssl: false
rules:
- pattern : 'kafka.rest<type=jetty-metrics>([^:]+):'
  name: "cp_kafka_rest_jetty_metrics_$1"
- pattern : 'kafka.rest<type=jersey-metrics>([^:]+):'
  name: "cp_kafka_rest_jersey_metrics_$1"
`

const kafkaCpKsqlServerJmxConfigmap = `
jmxUrl: service:jmx:rmi:///jndi/rmi://localhost:5555/jmxrmi
lowercaseOutputName: true
lowercaseOutputLabelNames: true
ssl: false
rules:
- pattern : 'io.confluent.ksql.metrics<type=ksql-engine-query-stats>([^:]+):'
  name: "cp_ksql_server_metrics_$1"
`
