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
	Context      k8s.Context
	StorageClass k8s.StorageClassName
	AccessMode   k8s.AccessMode
}

func (c *Kafka) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		c.Context,
		c.StorageClass,
		c.AccessMode,
	)
}
func (c *Kafka) Children() []world.Configuration {
	result := []world.Configuration{
		&k8s.NamespaceConfiguration{
			Context: c.Context,
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
	result = append(result, c.zookeeper()...)
	result = append(result, c.kafka()...)
	result = append(result, c.connect()...)
	result = append(result, c.rest()...)
	result = append(result, c.ksql()...)
	result = append(result, c.schemaRegistry()...)
	return result
}

func (c *Kafka) kafkaReplicas() k8s.Replicas {
	return 1
}

func (c *Kafka) zookeeperReplicas() k8s.Replicas {
	return 1
}

func (c *Kafka) connect() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/cp-kafka-connect",
		Tag:        "5.0.0",
	}
	return []world.Configuration{
		&build.CpKafkaConnect{
			Image: image,
		},
		&k8s.DeploymentConfiguration{
			Context: c.Context,
			Deployment: k8s.Deployment{
				ApiVersion: "apps/v1",
				Kind:       "Deployment",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka-cp-kafka-connect",
					Labels: k8s.Labels{
						"app":   "cp-kafka-connect",
						"chart": "cp-kafka-connect-0.1.0",
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
								k8s.Container{
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
										k8s.ContainerPort{
											ContainerPort: 5556,
										},
									},
									VolumeMounts: []k8s.ContainerMount{
										k8s.ContainerMount{
											Path: "/etc/jmx-kafka-connect",
											Name: "jmx-config",
										},
									},
								},
								k8s.Container{
									Name:  "cp-kafka-connect-server",
									Image: k8s.Image(image.String()),
									Env: []k8s.Env{
										k8s.Env{
											Name: "CONNECT_REST_ADVERTISED_HOST_NAME",
											ValueFrom: k8s.ValueFrom{
												FieldRef: k8s.FieldRef{
													FieldPath: "metadata.name",
												},
											},
										},
										k8s.Env{
											Name:  "CONNECT_CONFIG_STORAGE_REPLICATION_FACTOR",
											Value: c.kafkaReplicas().String(),
										},
										k8s.Env{
											Name:  "CONNECT_OFFSET_STORAGE_REPLICATION_FACTOR",
											Value: c.kafkaReplicas().String(),
										},
										k8s.Env{
											Name:  "CONNECT_STATUS_STORAGE_REPLICATION_FACTOR",
											Value: c.kafkaReplicas().String(),
										},
										k8s.Env{
											Name:  "CONNECT_PLUGIN_PATH",
											Value: "/usr/share/java",
										},
										k8s.Env{
											Name:  "CONNECT_BOOTSTRAP_SERVERS",
											Value: "PLAINTEXT://kafka-cp-kafka-headless:9092",
										},
										k8s.Env{
											Name:  "CONNECT_GROUP_ID",
											Value: "kafka",
										},
										k8s.Env{
											Name:  "CONNECT_CONFIG_STORAGE_TOPIC",
											Value: "kafka-cp-kafka-connect-config",
										},
										k8s.Env{
											Name:  "CONNECT_OFFSET_STORAGE_TOPIC",
											Value: "kafka-cp-kafka-connect-offset",
										},
										k8s.Env{
											Name:  "CONNECT_STATUS_STORAGE_TOPIC",
											Value: "kafka-cp-kafka-connect-status",
										},
										k8s.Env{
											Name:  "CONNECT_KEY_CONVERTER_SCHEMA_REGISTRY_URL",
											Value: "http://kafka-cp-schema-registry:8081",
										},
										k8s.Env{
											Name:  "CONNECT_VALUE_CONVERTER_SCHEMA_REGISTRY_URL",
											Value: "http://kafka-cp-schema-registry:8081",
										},
										k8s.Env{
											Name:  "CONNECT_KEY_CONVERTER",
											Value: "io.confluent.connect.avro.AvroConverter",
										},
										k8s.Env{
											Name:  "CONNECT_VALUE_CONVERTER",
											Value: "io.confluent.connect.avro.AvroConverter",
										},
										k8s.Env{
											Name:  "CONNECT_INTERNAL_KEY_CONVERTER",
											Value: "org.apache.kafka.connect.json.JsonConverter",
										},
										k8s.Env{
											Name:  "CONNECT_INTERNAL_VALUE_CONVERTER",
											Value: "org.apache.kafka.connect.json.JsonConverter",
										},
										k8s.Env{
											Name:  "KAFKA_JMX_PORT",
											Value: "5555",
										},
									},
									Ports: []k8s.ContainerPort{
										k8s.ContainerPort{
											ContainerPort: 8083,
											Name:          "kafka-connect",
											Protocol:      "TCP",
										},
										k8s.ContainerPort{
											ContainerPort: 5555,
											Name:          "jmx",
										},
									},
									ImagePullPolicy: "IfNotPresent",
								},
							},
							Volumes: []k8s.PodVolume{
								k8s.PodVolume{
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
			Context: c.Context,
			ConfigMap: k8s.ConfigMap{
				ApiVersion: "v1",
				Kind:       "ConfigMap",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka-cp-kafka-connect-jmx-configmap",
					Labels: k8s.Labels{
						"app":   "cp-kafka-connect",
						"chart": "cp-kafka-connect-0.1.0",
					},
				},
				Data: k8s.ConfigMapData{
					"jmx-kafka-connect-prometheus.yml": "jmxUrl: service:jmx:rmi:///jndi/rmi://localhost:5555/jmxrmi\nlowercaseOutputName: true\nlowercaseOutputLabelNames: true\nssl: false\nrules:\n- pattern : \"kafka.connect<type=connect-worker-metrics>([^:]+):\"\n  name: \"cp_kafka_connect_connect_worker_metrics_$1\"\n- pattern : \"kafka.connect<type=connect-metrics,client-id=([^:]+)><>([^:]+)\"\n  name: \"cp_kafka_connect_connect_metrics_$1_$2\"",
				},
			},
		},
		&k8s.ServiceConfiguration{
			Context: c.Context,
			Service: k8s.Service{
				ApiVersion: "v1",
				Kind:       "Service",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka-cp-kafka-connect",
					Labels: k8s.Labels{
						"app":   "cp-kafka-connect",
						"chart": "cp-kafka-connect-0.1.0",
					},
				},
				Spec: k8s.ServiceSpec{
					Ports: []k8s.ServicePort{
						k8s.ServicePort{
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
func (c *Kafka) kafka() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/cp-kafka",
		Tag:        "5.0.0",
	}
	return []world.Configuration{
		&build.CpKafka{
			Image: image,
		},
		&k8s.ServiceConfiguration{
			Context: c.Context,
			Service: k8s.Service{
				ApiVersion: "v1",
				Kind:       "Service",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka-cp-kafka-headless",
					Labels: k8s.Labels{
						"app":   "cp-kafka",
						"chart": "cp-kafka-0.1.0",
					},
				},
				Spec: k8s.ServiceSpec{
					Ports: []k8s.ServicePort{
						k8s.ServicePort{
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
		&k8s.ConfigMapConfiguration{
			Context: c.Context,
			ConfigMap: k8s.ConfigMap{
				ApiVersion: "v1",
				Kind:       "ConfigMap",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka-cp-kafka-jmx-configmap",
					Labels: k8s.Labels{
						"chart": "cp-kafka-0.1.0",
						"app":   "cp-kafka",
					},
				},
				Data: k8s.ConfigMapData{
					"jmx-kafka-prometheus.yml": "jmxUrl: service:jmx:rmi:///jndi/rmi://localhost:5555/jmxrmi\nlowercaseOutputName: true\nlowercaseOutputLabelNames: true\nssl: false\nrules:\n- pattern : kafka.server<type=ReplicaManager,name=(.+)><>(Value|OneMinuteRate)\n  name: \"cp_kafka_server_replicamanager_$1\"\n- pattern : kafka.controller<type=KafkaController,name=(.+)><>Value\n  name: \"cp_kafka_controller_kafkacontroller_$1\"\n- pattern : kafka.server<type=BrokerTopicMetrics,name=(.+)><>OneMinuteRate\n  name: \"cp_kafka_server_brokertopicmetrics_$1\"\n- pattern : kafka.network<type=RequestMetrics,name=RequestsPerSec,request=(.+)><>OneMinuteRate\n  name: \"cp_kafka_network_requestmetrics_requestspersec_$1\"\n- pattern : kafka.network<type=SocketServer,name=NetworkProcessorAvgIdlePercent><>Value\n  name: \"cp_kafka_network_socketserver_networkprocessoravgidlepercent\"\n- pattern : kafka.server<type=ReplicaFetcherManager,name=MaxLag,clientId=(.+)><>Value\n  name: \"cp_kafka_server_replicafetchermanager_maxlag_$1\"\n- pattern : kafka.server<type=KafkaRequestHandlerPool,name=RequestHandlerAvgIdlePercent><>OneMinuteRate\n  name: \"cp_kafka_kafkarequesthandlerpool_requesthandleravgidlepercent\"\n- pattern : kafka.controller<type=ControllerStats,name=(.+)><>OneMinuteRate\n  name: \"cp_kafka_controller_controllerstats_$1\"\n- pattern : kafka.server<type=SessionExpireListener,name=(.+)><>OneMinuteRate\n  name: \"cp_kafka_server_sessionexpirelistener_$1\"",
				},
			},
		},
		&k8s.StatefulSetConfiguration{
			Context: c.Context,
			StatefulSet: k8s.StatefulSet{
				ApiVersion: "apps/v1beta1",
				Kind:       "StatefulSet",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka-cp-kafka",
					Labels: k8s.Labels{
						"app":   "cp-kafka",
						"chart": "cp-kafka-0.1.0",
					},
				},
				Spec: k8s.StatefulSetSpec{
					ServiceName: "kafka-cp-kafka-headless",
					Replicas:    c.kafkaReplicas(),
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
							Containers: []k8s.Container{
								k8s.Container{
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
										k8s.ContainerPort{
											ContainerPort: 5556,
										},
									},
									VolumeMounts: []k8s.ContainerMount{
										k8s.ContainerMount{
											Path: "/etc/jmx-kafka",
											Name: "jmx-config",
										},
									},
								},
								k8s.Container{
									Name:  "cp-kafka-broker",
									Image: k8s.Image(image.String()),
									Command: []k8s.Command{
										"sh",
										"-exc",
										"export KAFKA_BROKER_ID=${HOSTNAME##*-} && \\\nexport KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://${POD_IP}:9092,EXTERNAL://${HOST_IP}:$((31090 + ${KAFKA_BROKER_ID})) && \\\nexec /etc/confluent/docker/run\n",
									},
									Env: []k8s.Env{
										k8s.Env{
											Name: "POD_IP",
											ValueFrom: k8s.ValueFrom{
												FieldRef: k8s.FieldRef{
													FieldPath: "status.podIP",
												},
											},
										},
										k8s.Env{
											Name: "HOST_IP",
											ValueFrom: k8s.ValueFrom{
												FieldRef: k8s.FieldRef{
													FieldPath: "status.hostIP",
												},
											},
										},
										k8s.Env{
											Name:  "KAFKA_HEAP_OPTS",
											Value: "-Xms512M -Xmx512M",
										},
										k8s.Env{
											Name:  "KAFKA_ZOOKEEPER_CONNECT",
											Value: "kafka-cp-zookeeper-headless:2181",
										},
										k8s.Env{
											Name:  "KAFKA_ADVERTISED_LISTENERS",
											Value: "EXTERNAL://${HOST_IP}:$((31090 + ${KAFKA_BROKER_ID}))",
										},
										k8s.Env{
											Name:  "KAFKA_LISTENER_SECURITY_PROTOCOL_MAP",
											Value: "PLAINTEXT:PLAINTEXT,EXTERNAL:PLAINTEXT",
										},
										k8s.Env{
											Name:  "KAFKA_LOG_DIRS",
											Value: "/opt/kafka/data/logs",
										},
										k8s.Env{
											Name:  "KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR",
											Value: c.kafkaReplicas().String(),
										},
										k8s.Env{
											Name:  "KAFKA_JMX_PORT",
											Value: "5555",
										},
									},
									Ports: []k8s.ContainerPort{
										k8s.ContainerPort{
											ContainerPort: 9092,
											Name:          "kafka",
										},
										k8s.ContainerPort{
											ContainerPort: 5555,
											Name:          "jmx",
										},
									},
									Resources: k8s.Resources{
										Limits:   k8s.ContainerResource{},
										Requests: k8s.ContainerResource{},
									},
									VolumeMounts: []k8s.ContainerMount{
										k8s.ContainerMount{
											Path: "/opt/kafka/data",
											Name: "datadir",
										},
									},
									ImagePullPolicy: "IfNotPresent",
								},
							},
							Volumes: []k8s.PodVolume{
								k8s.PodVolume{
									Name: "jmx-config",
									ConfigMap: k8s.PodVolumeConfigMap{
										Name: "kafka-cp-kafka-jmx-configmap",
									},
								},
							},
						},
					},
					VolumeClaimTemplates: []k8s.VolumeClaimTemplate{
						k8s.VolumeClaimTemplate{
							Metadata: k8s.Metadata{
								Name: "datadir",
								Annotations: map[string]string{
									"volume.alpha.kubernetes.io/storage-class": c.StorageClass.String(),
									"volume.beta.kubernetes.io/storage-class":  c.StorageClass.String(),
								},
							},
							Spec: k8s.VolumeClaimTemplatesSpec{
								AccessModes: []k8s.AccessMode{c.AccessMode},
								Resources: k8s.VolumeClaimTemplatesSpecResources{
									Requests: k8s.VolumeClaimTemplatesSpecResourcesRequests{
										Storage: "5Gi",
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
			Context: c.Context,
			Service: k8s.Service{
				ApiVersion: "v1",
				Kind:       "Service",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka-cp-kafka",
					Labels: k8s.Labels{
						"app":   "cp-kafka",
						"chart": "cp-kafka-0.1.0",
					},
				},
				Spec: k8s.ServiceSpec{
					Ports: []k8s.ServicePort{
						k8s.ServicePort{
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
func (c *Kafka) rest() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/cp-kafka-rest",
		Tag:        "5.0.0",
	}
	return []world.Configuration{
		&build.CpKafkaRest{
			Image: image,
		},
		&k8s.DeploymentConfiguration{
			Context: c.Context,
			Deployment: k8s.Deployment{
				ApiVersion: "apps/v1",
				Kind:       "Deployment",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka-cp-kafka-rest",
					Labels: k8s.Labels{
						"app":   "cp-kafka-rest",
						"chart": "cp-kafka-rest-0.1.0",
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
							Containers: []k8s.Container{
								k8s.Container{
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
										k8s.ContainerPort{
											ContainerPort: 5556,
										},
									},
									VolumeMounts: []k8s.ContainerMount{
										k8s.ContainerMount{
											Path: "/etc/jmx-kafka-rest",
											Name: "jmx-config",
										},
									},
								},
								k8s.Container{
									Name:  "cp-kafka-rest-server",
									Image: k8s.Image(image.String()),
									Env: []k8s.Env{
										k8s.Env{
											Name: "KAFKA_REST_HOST_NAME",
											ValueFrom: k8s.ValueFrom{
												SecretKeyRef: k8s.SecretKeyRef{},
												FieldRef: k8s.FieldRef{
													FieldPath: "metadata.name",
												},
											},
										},
										k8s.Env{
											Name:  "KAFKA_REST_ZOOKEEPER_CONNECT",
											Value: "kafka-cp-zookeeper-headless:2181",
										},
										k8s.Env{
											Name:  "KAFKA_REST_SCHEMA_REGISTRY_URL",
											Value: "http://kafka-cp-schema-registry:8081",
										},
										k8s.Env{
											Name:  "KAFKA_REST_JMX_PORT",
											Value: "5555",
										},
									},
									Ports: []k8s.ContainerPort{
										k8s.ContainerPort{
											ContainerPort: 8082,
											Name:          "rest-proxy",
											Protocol:      "TCP",
										},
										k8s.ContainerPort{
											ContainerPort: 5555,
											Name:          "jmx",
										},
									},
									ImagePullPolicy: "IfNotPresent",
								},
							},
							Volumes: []k8s.PodVolume{
								k8s.PodVolume{
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
			Context: c.Context,
			ConfigMap: k8s.ConfigMap{
				ApiVersion: "v1",
				Kind:       "ConfigMap",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka-cp-kafka-rest-jmx-configmap",
					Labels: k8s.Labels{
						"app":   "cp-kafka-rest",
						"chart": "cp-kafka-rest-0.1.0",
					},
					Annotations: k8s.Annotations(nil),
				},
				Data: k8s.ConfigMapData{
					"jmx-kafka-rest-prometheus.yml": "jmxUrl: service:jmx:rmi:///jndi/rmi://localhost:5555/jmxrmi\nlowercaseOutputName: true\nlowercaseOutputLabelNames: true\nssl: false\nrules:\n- pattern : 'kafka.rest<type=jetty-metrics>([^:]+):'\n  name: \"cp_kafka_rest_jetty_metrics_$1\"\n- pattern : 'kafka.rest<type=jersey-metrics>([^:]+):'\n  name: \"cp_kafka_rest_jersey_metrics_$1\"",
				},
			},
		},
		&k8s.ServiceConfiguration{
			Context: c.Context,
			Service: k8s.Service{
				ApiVersion: "v1",
				Kind:       "Service",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka-cp-kafka-rest",
					Labels: k8s.Labels{
						"app":   "cp-kafka-rest",
						"chart": "cp-kafka-rest-0.1.0",
					},
					Annotations: k8s.Annotations(nil),
				},
				Spec: k8s.ServiceSpec{
					Ports: []k8s.ServicePort{
						k8s.ServicePort{
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
func (c *Kafka) ksql() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/cp-ksql-server",
		Tag:        "5.0.0",
	}
	return []world.Configuration{
		&build.CpKafkaKsql{
			Image: image,
		},
		&k8s.DeploymentConfiguration{
			Context: c.Context,
			Deployment: k8s.Deployment{
				ApiVersion: "apps/v1",
				Kind:       "Deployment",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka-cp-ksql-server",
					Labels: k8s.Labels{
						"app":   "cp-ksql-server",
						"chart": "cp-ksql-server-0.1.0",
					},
					Annotations: k8s.Annotations(nil)},
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
							Tolerations: []k8s.Toleration(nil),
							Containers: []k8s.Container{
								k8s.Container{
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
									Args: []k8s.Arg(nil),
									Env:  []k8s.Env(nil),
									Ports: []k8s.ContainerPort{
										k8s.ContainerPort{
											ContainerPort: 5556,
											HostPort:      0,
										},
									},
									Resources: k8s.Resources{
										Limits:   k8s.ContainerResource{},
										Requests: k8s.ContainerResource{},
									},
									VolumeMounts: []k8s.ContainerMount{
										k8s.ContainerMount{
											Path:     "/etc/jmx-ksql-server",
											Name:     "jmx-config",
											ReadOnly: false},
									},
									ReadinessProbe: k8s.Probe{
										Exec: k8s.Exec{
											Command: []k8s.Command(nil)},
										HttpGet: k8s.HttpGet{
											Port: 0,
										},
										TcpSocket: k8s.TcpSocket{
											Port: 0},
										InitialDelaySeconds: 0,
										SuccessThreshold:    0,
										FailureThreshold:    0,
										TimeoutSeconds:      0,
										PeriodSeconds:       0},
									LivenessProbe: k8s.Probe{
										Exec: k8s.Exec{
											Command: []k8s.Command(nil)},
										HttpGet: k8s.HttpGet{
											Port: 0,
										},
										TcpSocket: k8s.TcpSocket{
											Port: 0},
										InitialDelaySeconds: 0,
										SuccessThreshold:    0,
										FailureThreshold:    0,
										TimeoutSeconds:      0,
										PeriodSeconds:       0},
									SecurityContext: k8s.SecurityContext{
										AllowPrivilegeEscalation: false,
										ReadOnlyRootFilesystem:   false,
										Privileged:               false,
										RunAsUser:                0,
										FsGroup:                  0,
										Capabilities:             k8s.SecurityContextCapabilities(nil)},
								},
								k8s.Container{
									Name:    "cp-ksql-server",
									Image:   k8s.Image(image.String()),
									Command: []k8s.Command(nil),
									Args:    []k8s.Arg(nil),
									Env: []k8s.Env{
										k8s.Env{
											Name:  "KSQL_BOOTSTRAP_SERVERS",
											Value: "PLAINTEXT://kafka-cp-kafka-headless:9092",
										},
										k8s.Env{
											Name:  "KSQL_KSQL_SERVICE_ID",
											Value: "kafka",
										},
										k8s.Env{
											Name:  "KSQL_LISTENERS",
											Value: "http://0.0.0.0:8088",
										},
										k8s.Env{
											Name:  "KSQL_JMX_PORT",
											Value: "5555",
										},
									},
									Ports: []k8s.ContainerPort{
										k8s.ContainerPort{
											ContainerPort: 8088,
											Name:          "server",
											Protocol:      "TCP",
										},
										k8s.ContainerPort{
											ContainerPort: 5555,
											Name:          "jmx",
										},
									},
									ImagePullPolicy: "IfNotPresent",
								},
							},
							Volumes: []k8s.PodVolume{
								k8s.PodVolume{
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
			Context: c.Context,
			ConfigMap: k8s.ConfigMap{
				ApiVersion: "v1",
				Kind:       "ConfigMap",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka-cp-ksql-server-jmx-configmap",
					Labels: k8s.Labels{
						"app":   "cp-ksql-server",
						"chart": "cp-ksql-server-0.1.0",
					},
				},
				Data: k8s.ConfigMapData{
					"jmx-ksql-server-prometheus.yml": "jmxUrl: service:jmx:rmi:///jndi/rmi://localhost:5555/jmxrmi\nlowercaseOutputName: true\nlowercaseOutputLabelNames: true\nssl: false\nrules:\n- pattern : 'io.confluent.ksql.metrics<type=ksql-engine-query-stats>([^:]+):'\n  name: \"cp_ksql_server_metrics_$1\"",
				},
			},
		},
		&k8s.ConfigMapConfiguration{
			Context: c.Context,
			ConfigMap: k8s.ConfigMap{
				ApiVersion: "v1",
				Kind:       "ConfigMap",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka-cp-ksql-server-ksql-queries-configmap",
					Labels: k8s.Labels{
						"app":   "cp-ksql-server",
						"chart": "cp-ksql-server-0.1.0",
					},
				},
				Data: k8s.ConfigMapData{
					"queries.sql": "-- From http://docs.confluent.io/current/ksql/docs/tutorials/basics-docker.html#create-a-stream-and-table\n\n-- Create a stream pageviews_original from the Kafka topic pageviews, specifying the value_format of DELIMITED\nCREATE STREAM pageviews_original (viewtime bigint, userid varchar, pageid varchar) WITH (kafka_topic='pageviews', value_format='DELIMITED');\n\n-- Create a table users_original from the Kafka topic users, specifying the value_format of JSON\nCREATE TABLE users_original (registertime BIGINT, gender VARCHAR, regionid VARCHAR, userid VARCHAR) WITH (kafka_topic='users', value_format='JSON', key = 'userid');\n\n-- Create a persistent query by using the CREATE STREAM keywords to precede the SELECT statement\nCREATE STREAM pageviews_enriched AS SELECT users_original.userid AS userid, pageid, regionid, gender FROM pageviews_original LEFT JOIN users_original ON pageviews_original.userid = users_original.userid;\n\n-- Create a new persistent query where a condition limits the streams content, using WHERE\nCREATE STREAM pageviews_female AS SELECT * FROM pageviews_enriched WHERE gender = 'FEMALE';\n\n-- Create a new persistent query where another condition is met, using LIKE\nCREATE STREAM pageviews_female_like_89 WITH (kafka_topic='pageviews_enriched_r8_r9') AS SELECT * FROM pageviews_female WHERE regionid LIKE '%_8' OR regionid LIKE '%_9';\n\n-- Create a new persistent query that counts the pageviews for each region and gender combination in a tumbling window of 30 seconds when the count is greater than one\nCREATE TABLE pageviews_regions WITH (VALUE_FORMAT='avro') AS SELECT gender, regionid , COUNT(*) AS numusers FROM pageviews_enriched WINDOW TUMBLING (size 30 second) GROUP BY gender, regionid HAVING COUNT(*) > 1;",
				},
			},
		},
		&k8s.ServiceConfiguration{
			Context: c.Context,
			Service: k8s.Service{
				ApiVersion: "v1",
				Kind:       "Service",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka-cp-ksql-server",
					Labels: k8s.Labels{
						"app":   "cp-ksql-server",
						"chart": "cp-ksql-server-0.1.0",
					},
				},
				Spec: k8s.ServiceSpec{
					Ports: []k8s.ServicePort{
						k8s.ServicePort{
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
func (c *Kafka) schemaRegistry() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/cp-schema-registry",
		Tag:        "5.0.0",
	}
	return []world.Configuration{
		&build.CpKafkaSchemaRegistry{
			Image: image,
		},
		&k8s.DeploymentConfiguration{
			Context: c.Context,
			Deployment: k8s.Deployment{
				ApiVersion: "apps/v1",
				Kind:       "Deployment",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka-cp-schema-registry",
					Labels: k8s.Labels{
						"app":   "cp-schema-registry",
						"chart": "cp-schema-registry-0.1.0",
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
							Containers: []k8s.Container{
								k8s.Container{
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
										k8s.ContainerPort{
											ContainerPort: 5556,
										},
									},
									VolumeMounts: []k8s.ContainerMount{
										k8s.ContainerMount{
											Path: "/etc/jmx-schema-registry",
											Name: "jmx-config",
										},
									},
								},
								k8s.Container{
									Name:  "cp-schema-registry-server",
									Image: k8s.Image(image.String()),
									Env: []k8s.Env{
										k8s.Env{
											Name: "SCHEMA_REGISTRY_HOST_NAME",
											ValueFrom: k8s.ValueFrom{
												FieldRef: k8s.FieldRef{
													FieldPath: "metadata.name",
												},
											},
										},
										k8s.Env{
											Name:  "SCHEMA_REGISTRY_KAFKASTORE_BOOTSTRAP_SERVERS",
											Value: "PLAINTEXT://kafka-cp-kafka-headless:9092",
										},
										k8s.Env{
											Name:  "SCHEMA_REGISTRY_KAFKASTORE_GROUP_ID",
											Value: "kafka",
										},
										k8s.Env{
											Name:  "SCHEMA_REGISTRY_MASTER_ELIGIBILITY",
											Value: "true",
										},
										k8s.Env{
											Name:  "JMX_PORT",
											Value: "5555",
										},
									},
									Ports: []k8s.ContainerPort{
										k8s.ContainerPort{
											ContainerPort: 8081,
											Name:          "schema-registry",
											Protocol:      "TCP",
										},
										k8s.ContainerPort{
											ContainerPort: 5555,
											Name:          "jmx",
										},
									},
									ImagePullPolicy: "IfNotPresent",
								},
							},
							Volumes: []k8s.PodVolume{
								k8s.PodVolume{
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
			Context: c.Context,
			ConfigMap: k8s.ConfigMap{
				ApiVersion: "v1",
				Kind:       "ConfigMap",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka-cp-schema-registry-jmx-configmap",
					Labels: k8s.Labels{
						"app":   "cp-schema-registry",
						"chart": "cp-schema-registry-0.1.0",
					},
				},
				Data: k8s.ConfigMapData{
					"jmx-schema-registry-prometheus.yml": "jmxUrl: service:jmx:rmi:///jndi/rmi://localhost:5555/jmxrmi\nlowercaseOutputName: true\nlowercaseOutputLabelNames: true\nssl: false\nrules:\n- pattern : 'kafka.schema.registry<type=jetty-metrics>([^:]+):'\n  name: \"cp_kafka_schema_registry_jetty_metrics_$1\"\n- pattern : 'kafka.schema.registry<type=master-slave-role>([^:]+):'\n  name: \"cp_kafka_schema_registry_master_slave_role\"\n- pattern : 'kafka.schema.registry<type=jersey-metrics>([^:]+):'\n  name: \"cp_kafka_schema_registry_jersey_metrics_$1\"",
				},
			},
		},
		&k8s.ServiceConfiguration{
			Context: c.Context,
			Service: k8s.Service{
				ApiVersion: "v1",
				Kind:       "Service",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka-cp-schema-registry",
					Labels: k8s.Labels{
						"app":   "cp-schema-registry",
						"chart": "cp-schema-registry-0.1.0",
					},
				},
				Spec: k8s.ServiceSpec{
					Ports: []k8s.ServicePort{
						k8s.ServicePort{
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
func (c *Kafka) zookeeper() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/cp-zookeeper",
		Tag:        "5.0.0",
	}
	var zookeeperServerLists []string
	for i := k8s.Replicas(0); i < c.zookeeperReplicas(); i++ {
		addr := fmt.Sprintf("kafka-cp-zookeeper-%d.kafka-cp-zookeeper-headless.default.svc.cluster.local:2888:3888", i)
		zookeeperServerLists = append(zookeeperServerLists, addr)
	}
	return []world.Configuration{
		&build.CpZookeeper{
			Image: image,
		},
		&k8s.ServiceConfiguration{
			Context: c.Context,
			Service: k8s.Service{
				ApiVersion: "v1",
				Kind:       "Service",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka-cp-zookeeper-headless",
					Labels: k8s.Labels{
						"chart": "cp-zookeeper-0.1.0",
						"app":   "cp-zookeeper",
					},
				},
				Spec: k8s.ServiceSpec{
					Ports: []k8s.ServicePort{
						k8s.ServicePort{
							Name: "server",
							Port: 2888,
						},
						k8s.ServicePort{
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
			Context: c.Context,
			ConfigMap: k8s.ConfigMap{
				ApiVersion: "v1",
				Kind:       "ConfigMap",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka-cp-zookeeper-jmx-configmap",
					Labels: k8s.Labels{
						"app":   "cp-zookeeper",
						"chart": "cp-zookeeper-0.1.0",
					},
				},
				Data: k8s.ConfigMapData{
					"jmx-zookeeper-prometheus.yml": "jmxUrl: service:jmx:rmi:///jndi/rmi://localhost:5555/jmxrmi\nlowercaseOutputName: true\nlowercaseOutputLabelNames: true\nssl: false\nrules:\n- pattern: \"org.apache.ZooKeeperService<name0=ReplicatedServer_id(\\\\d+)><>(\\\\w+)\"\n  name: \"cp_zookeeper_$2\"\n- pattern: \"org.apache.ZooKeeperService<name0=ReplicatedServer_id(\\\\d+),name1=replica.(\\\\d+)><>(\\\\w+)\"\n  name: \"cp_zookeeper_$3\"\n  labels:\n    replicaId: \"$2\"\n- pattern: \"org.apache.ZooKeeperService<name0=ReplicatedServer_id(\\\\d+),name1=replica.(\\\\d+),name2=(\\\\w+)><>(\\\\w+)\"\n  name: \"cp_zookeeper_$4\"\n  labels:\n    replicaId: \"$2\"\n    memberType: \"$3\"\n- pattern: \"org.apache.ZooKeeperService<name0=ReplicatedServer_id(\\\\d+),name1=replica.(\\\\d+),name2=(\\\\w+),name3=(\\\\w+)><>(\\\\w+)\"\n  name: \"cp_zookeeper_$4_$5\"\n  labels:\n    replicaId: \"$2\"\n    memberType: \"$3\"",
				},
			},
		},
		&k8s.PodDisruptionBudgetConfiguration{
			Context: c.Context,
			PodDisruptionBudget: k8s.PodDisruptionBudget{
				ApiVersion: "policy/v1beta1",
				Kind:       "PodDisruptionBudget",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka-cp-zookeeper-pdb",
					Labels: k8s.Labels{
						"app":   "cp-zookeeper",
						"chart": "cp-zookeeper-0.1.0",
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
			Context: c.Context,
			StatefulSet: k8s.StatefulSet{
				ApiVersion: "apps/v1beta1",
				Kind:       "StatefulSet",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka-cp-zookeeper",
					Labels: k8s.Labels{
						"app":   "cp-zookeeper",
						"chart": "cp-zookeeper-0.1.0",
					},
				},
				Spec: k8s.StatefulSetSpec{
					ServiceName: "kafka-cp-zookeeper-headless",
					Replicas:    k8s.Replicas(c.zookeeperReplicas()),
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
							Containers: []k8s.Container{
								k8s.Container{
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
										k8s.ContainerPort{
											ContainerPort: 5556,
										},
									},
									VolumeMounts: []k8s.ContainerMount{
										k8s.ContainerMount{
											Path: "/etc/jmx-zookeeper",
											Name: "jmx-config",
										},
									},
								},
								k8s.Container{
									Name:  "cp-zookeeper-server",
									Image: k8s.Image(image.String()),
									Command: []k8s.Command{
										"bash",
										"-c",
										"ZOOKEEPER_SERVER_ID=$((${HOSTNAME##*-}+1)) && /etc/confluent/docker/run",
									},
									Env: []k8s.Env{
										k8s.Env{
											Name:  "KAFKA_HEAP_OPTS",
											Value: "-Xms512M -Xmx512M",
										},
										k8s.Env{
											Name:  "KAFKA_JMX_PORT",
											Value: "5555",
										},
										k8s.Env{
											Name:  "ZOOKEEPER_TICK_TIME",
											Value: "2000",
										},
										k8s.Env{
											Name:  "ZOOKEEPER_SYNC_LIMIT",
											Value: "5",
										},
										k8s.Env{
											Name:  "ZOOKEEPER_INIT_LIMIT",
											Value: "10",
										},
										k8s.Env{
											Name:  "ZOOKEEPER_MAX_CLIENT_CNXNS",
											Value: "60",
										},
										k8s.Env{
											Name:  "ZOOKEEPER_AUTOPURGE_SNAP_RETAIN_COUNT",
											Value: "3",
										},
										k8s.Env{
											Name:  "ZOOKEEPER_AUTOPURGE_PURGE_INTERVAL",
											Value: "24",
										},
										k8s.Env{
											Name:  "ZOOKEEPER_CLIENT_PORT",
											Value: "2181",
										},
										k8s.Env{
											Name:  "ZOOKEEPER_SERVERS",
											Value: strings.Join(zookeeperServerLists, ";"),
										},
										k8s.Env{
											Name: "ZOOKEEPER_SERVER_ID",
											ValueFrom: k8s.ValueFrom{
												FieldRef: k8s.FieldRef{
													FieldPath: "metadata.name",
												},
											},
										},
									},
									Ports: []k8s.ContainerPort{
										k8s.ContainerPort{
											ContainerPort: 2181,
											Name:          "client",
										},
										k8s.ContainerPort{
											ContainerPort: 2888,
											Name:          "server",
										},
										k8s.ContainerPort{
											ContainerPort: 3888,
											Name:          "leader-election",
										},
										k8s.ContainerPort{
											ContainerPort: 5555,
											Name:          "jmx",
										},
									},
									VolumeMounts: []k8s.ContainerMount{
										k8s.ContainerMount{
											Path: "/var/lib/zookeeper/data",
											Name: "datadir",
										},
										k8s.ContainerMount{
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
									ImagePullPolicy: "IfNotPresent",
								},
							},
							Volumes: []k8s.PodVolume{
								k8s.PodVolume{
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
									"volume.alpha.kubernetes.io/storage-class": c.StorageClass.String(),
									"volume.beta.kubernetes.io/storage-class":  c.StorageClass.String(),
								},
							},
							Spec: k8s.VolumeClaimTemplatesSpec{
								AccessModes: []k8s.AccessMode{c.AccessMode},
								Resources: k8s.VolumeClaimTemplatesSpecResources{
									Requests: k8s.VolumeClaimTemplatesSpecResourcesRequests{
										Storage: "5Gi",
									},
								},
							},
						},
						{
							Metadata: k8s.Metadata{
								Name: "datalogdir",
								Annotations: map[string]string{
									"volume.alpha.kubernetes.io/storage-class": c.StorageClass.String(),
									"volume.beta.kubernetes.io/storage-class":  c.StorageClass.String(),
								},
							},
							Spec: k8s.VolumeClaimTemplatesSpec{
								AccessModes: []k8s.AccessMode{c.AccessMode},
								Resources: k8s.VolumeClaimTemplatesSpecResources{
									Requests: k8s.VolumeClaimTemplatesSpecResourcesRequests{
										Storage: "5Gi",
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
			Context: c.Context,
			Service: k8s.Service{
				ApiVersion: "v1",
				Kind:       "Service",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka-cp-zookeeper",
					Labels: k8s.Labels{
						"chart": "cp-zookeeper-0.1.0",
						"app":   "cp-zookeeper",
					},
				},
				Spec: k8s.ServiceSpec{
					Ports: []k8s.ServicePort{
						k8s.ServicePort{
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
func (c *Kafka) Applier() (world.Applier, error) {
	return nil, nil
}
