package app

import (
	"context"

	"github.com/bborbe/world/pkg/k8s"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Kafka struct {
	Cluster cluster.Cluster
}

func (k *Kafka) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		k.Cluster,
	)
}

func (k *Kafka) Children() []world.Configuration {
	result := []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   k.Cluster.Context,
			Namespace: "kafka",
		},
	}
	result = append(result, k.zookeeper()...)
	result = append(result, k.kafka()...)
	return result
}

func (k *Kafka) zookeeper() []world.Configuration {
	clientPort := deployer.Port{
		Port:     2181,
		Protocol: "TCP",
		Name:     "client",
	}
	serverPort := deployer.Port{
		Port:     2888,
		Protocol: "TCP",
		Name:     "server",
	}
	leaderElectionPort := deployer.Port{
		Port:     3888,
		Protocol: "TCP",
		Name:     "leader-election",
	}

	image := docker.Image{
		Repository: "bborbe/zookeeper",
		Tag:        "master",
	}
	replicas := k8s.Replicas(1)
	return []world.Configuration{
		&k8s.ServiceConfiguration{
			Context: k.Cluster.Context,
			Service: k8s.Service{
				ApiVersion: "v1",
				Kind:       "Service",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "zookeeper",
				},
				Spec: k8s.ServiceSpec{
					Ports: []k8s.ServicePort{
						serverPort.ServicePort(),
						leaderElectionPort.ServicePort(),
					},
					Selector: k8s.ServiceSelector{
						"app": "zookeeper",
					},
				},
			},
		},
		&k8s.StatefulSetConfiguration{
			Context: k.Cluster.Context,
			Requirements: []world.Configuration{
				&build.Zookeeper{
					Image: image,
				},
			},
			StatefulSet: k8s.StatefulSet{
				ApiVersion: "apps/v1beta1",
				Kind:       "StatefulSet",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "zookeeper",
				},
				Spec: k8s.StatefulSetSpec{
					Selector: k8s.Selector{
						MatchLabels: k8s.Labels{
							"app": "zookeeper",
						},
					},
					ServiceName: "zookeeper",
					Replicas:    replicas,
					Template: k8s.PodTemplate{
						Metadata: k8s.Metadata{
							Labels: map[string]string{
								"app": "zookeeper",
							},
						},
						Spec: k8s.PodSpec{
							Containers: []k8s.Container{
								{
									Name:            "kafka",
									ImagePullPolicy: "Always",
									Image:           k8s.Image(image.String()),
									Resources: k8s.Resources{
										Limits: k8s.ContainerResource{
											Cpu:    "2000m",
											Memory: "1000Mi",
										},
										Requests: k8s.ContainerResource{
											Cpu:    "100m",
											Memory: "500Mi",
										},
									},
									Ports: []k8s.ContainerPort{
										clientPort.ContainerPort(),
										serverPort.ContainerPort(),
										leaderElectionPort.ContainerPort(),
									},
									Command: []k8s.Command{
										"sh",
										"-c",
										"zkGenConfig.sh && zkServer.sh start-foreground",
									},
									Env: []k8s.Env{
										{
											Name:  "ZK_REPLICAS",
											Value: replicas.String(),
										},
										{
											Name:  "ZK_HEAP_SIZE",
											Value: "1G",
										},
										{
											Name:  "ZK_TICK_TIME",
											Value: "2000",
										},
										{
											Name:  "ZK_INIT_LIMIT",
											Value: "10",
										},
										{
											Name:  "ZK_SYNC_LIMIT",
											Value: "5",
										},
										{
											Name:  "ZK_MAX_CLIENT_CNXNS",
											Value: "60",
										},
										{
											Name:  "ZK_SNAP_RETAIN_COUNT",
											Value: "3",
										},
										{
											Name:  "ZK_PURGE_INTERVAL",
											Value: "0",
										},
										{
											Name:  "ZK_CLIENT_PORT",
											Value: clientPort.Port.String(),
										},
										{
											Name:  "ZK_SERVER_PORT",
											Value: serverPort.Port.String(),
										},
										{
											Name:  "ZK_ELECTION_PORT",
											Value: leaderElectionPort.Port.String(),
										},
									},
									ReadinessProbe: k8s.Probe{
										Exec: k8s.Exec{
											Command: []k8s.Command{"zkOk.sh"},
										},
										InitialDelaySeconds: 10,
										TimeoutSeconds:      5,
									},
									LivenessProbe: k8s.Probe{
										Exec: k8s.Exec{
											Command: []k8s.Command{"zkOk.sh"},
										},
										InitialDelaySeconds: 10,
										TimeoutSeconds:      5,
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

func (k *Kafka) kafka() []world.Configuration {
	port := deployer.Port{
		Port:     9093,
		Protocol: "TCP",
		Name:     "server",
	}
	image := docker.Image{
		Repository: "bborbe/kafka",
		Tag:        "1.1.1",
	}
	return []world.Configuration{
		&k8s.StatefulSetConfiguration{
			Context: k.Cluster.Context,
			Requirements: []world.Configuration{
				&build.Kafka{
					Image: image,
				},
			},
			StatefulSet: k8s.StatefulSet{
				ApiVersion: "apps/v1beta1",
				Kind:       "StatefulSet",
				Metadata: k8s.Metadata{
					Namespace: "kafka",
					Name:      "kafka",
				},
				Spec: k8s.StatefulSetSpec{
					Selector: k8s.Selector{
						MatchLabels: k8s.Labels{
							"app": "kafka",
						},
					},
					ServiceName: "kafka",
					Replicas:    1,
					Template: k8s.PodTemplate{
						Metadata: k8s.Metadata{
							Labels: map[string]string{
								"app": "kafka",
							},
						},
						Spec: k8s.PodSpec{
							//SecurityContext: k8s.SecurityContext{
							//	RunAsUser: 1000,
							//	FsGroup:   1000,
							//},
							Containers: []k8s.Container{
								{
									Name:            "kafka",
									ImagePullPolicy: "Always",
									Image:           k8s.Image(image.String()),
									Resources: k8s.Resources{
										Limits: k8s.ContainerResource{
											Cpu:    "2000m",
											Memory: "1000Mi",
										},
										Requests: k8s.ContainerResource{
											Cpu:    "100m",
											Memory: "500Mi",
										},
									},
									Ports: []k8s.ContainerPort{
										port.ContainerPort(),
									},
									Command: []k8s.Command{
										"sh",
										"-c",
										kafkaCommand,
									},
									Env: []k8s.Env{
										{
											Name:  "KAFKA_HEAP_OPTS",
											Value: "-Xmx512M -Xms512M",
										},
										{
											Name:  "KAFKA_OPTS",
											Value: "-Dlogging.level=INFO",
										},
									},
									//VolumeMounts: []k8s.ContainerMount{
									//	{
									//		Name: "data",
									//		Path: "/var/lib/kafka",
									//	},
									//},
									ReadinessProbe: k8s.Probe{
										Exec: k8s.Exec{
											Command: []k8s.Command{
												"sh",
												"-c",
												"/opt/kafka/bin/kafka-broker-api-versions.sh --bootstrap-server=localhost:9093",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   k.Cluster.Context,
			Namespace: "kafka",
			Name:      "kafka",
			Ports:     []deployer.Port{port},
		},
	}
}

func (k *Kafka) Applier() (world.Applier, error) {
	return nil, nil
}

const kafkaCommand = `exec kafka-server-start.sh /opt/kafka/config/server.properties \
--override broker.id=${HOSTNAME##*-} \
--override listeners=PLAINTEXT://:9093 \
--override zookeeper.connect=zookeeper-0.zookeeper.kafka.svc.cluster.local:2181 \
--override log.dir=/var/lib/kafka \
--override auto.create.topics.enable=true \
--override auto.leader.rebalance.enable=true \
--override background.threads=10 \
--override compression.type=producer \
--override delete.topic.enable=false \
--override leader.imbalance.check.interval.seconds=300 \
--override leader.imbalance.per.broker.percentage=10 \
--override log.flush.interval.messages=9223372036854775807 \
--override log.flush.offset.checkpoint.interval.ms=60000 \
--override log.flush.scheduler.interval.ms=9223372036854775807 \
--override log.retention.bytes=-1 \
--override log.retention.hours=168 \
--override log.roll.hours=168 \
--override log.roll.jitter.hours=0 \
--override log.segment.bytes=1073741824 \
--override log.segment.delete.delay.ms=60000 \
--override message.max.bytes=1000012 \
--override min.insync.replicas=1 \
--override num.io.threads=8 \
--override num.network.threads=3 \
--override num.recovery.threads.per.data.dir=1 \
--override num.replica.fetchers=1 \
--override offset.metadata.max.bytes=4096 \
--override offsets.commit.required.acks=-1 \
--override offsets.commit.timeout.ms=5000 \
--override offsets.load.buffer.size=5242880 \
--override offsets.retention.check.interval.ms=600000 \
--override offsets.retention.minutes=1440 \
--override offsets.topic.compression.codec=0 \
--override offsets.topic.num.partitions=50 \
--override offsets.topic.replication.factor=3 \
--override offsets.topic.segment.bytes=104857600 \
--override queued.max.requests=500 \
--override quota.consumer.default=9223372036854775807 \
--override quota.producer.default=9223372036854775807 \
--override replica.fetch.min.bytes=1 \
--override replica.fetch.wait.max.ms=500 \
--override replica.high.watermark.checkpoint.interval.ms=5000 \
--override replica.lag.time.max.ms=10000 \
--override replica.socket.receive.buffer.bytes=65536 \
--override replica.socket.timeout.ms=30000 \
--override request.timeout.ms=30000 \
--override socket.receive.buffer.bytes=102400 \
--override socket.request.max.bytes=104857600 \
--override socket.send.buffer.bytes=102400 \
--override unclean.leader.election.enable=true \
--override zookeeper.session.timeout.ms=6000 \
--override zookeeper.set.acl=false \
--override broker.id.generation.enable=true \
--override connections.max.idle.ms=600000 \
--override controlled.shutdown.enable=true \
--override controlled.shutdown.max.retries=3 \
--override controlled.shutdown.retry.backoff.ms=5000 \
--override controller.socket.timeout.ms=30000 \
--override default.replication.factor=1 \
--override fetch.purgatory.purge.interval.requests=1000 \
--override group.max.session.timeout.ms=300000 \
--override group.min.session.timeout.ms=6000 \
--override inter.broker.protocol.version=1.1-IV0 \
--override log.cleaner.backoff.ms=15000 \
--override log.cleaner.dedupe.buffer.size=134217728 \
--override log.cleaner.delete.retention.ms=86400000 \
--override log.cleaner.enable=true \
--override log.cleaner.io.buffer.load.factor=0.9 \
--override log.cleaner.io.buffer.size=524288 \
--override log.cleaner.io.max.bytes.per.second=1.7976931348623157E308 \
--override log.cleaner.min.cleanable.ratio=0.5 \
--override log.cleaner.min.compaction.lag.ms=0 \
--override log.cleaner.threads=1 \
--override log.cleanup.policy=delete \
--override log.index.interval.bytes=4096 \
--override log.index.size.max.bytes=10485760 \
--override log.message.timestamp.difference.max.ms=9223372036854775807 \
--override log.message.timestamp.type=CreateTime \
--override log.preallocate=false \
--override log.retention.check.interval.ms=300000 \
--override max.connections.per.ip=2147483647 \
--override num.partitions=1 \
--override producer.purgatory.purge.interval.requests=1000 \
--override replica.fetch.backoff.ms=1000 \
--override replica.fetch.max.bytes=1048576 \
--override replica.fetch.response.max.bytes=10485760 \
--override reserved.broker.max.id=1000`
