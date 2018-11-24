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

type ErpNext struct {
	Context              k8s.Context
	NfsServer            k8s.PodNfsServer
	Domain               k8s.IngressHost
	DatabaseRootPassword deployer.SecretValue
	DatabaseName         deployer.SecretValue
	DatabasePassword     deployer.SecretValue
	AdminPassword        deployer.SecretValue
}

func (e *ErpNext) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		e.Context,
		e.NfsServer,
		e.Domain,
		e.DatabaseRootPassword,
		e.DatabaseName,
		e.DatabasePassword,
		e.AdminPassword,
	)
}

func (e *ErpNext) Applier() (world.Applier, error) {
	return nil, nil
}

func (e *ErpNext) Children() []world.Configuration {
	configurations := []world.Configuration{
		&k8s.NamespaceConfiguration{
			Context: e.Context,
			Namespace: k8s.Namespace{
				ApiVersion: "v1",
				Kind:       "Namespace",
				Metadata: k8s.Metadata{
					Namespace: "erpnext",
					Name:      "erpnext",
				},
			},
		},
	}
	configurations = append(configurations, e.redisCache()...)
	configurations = append(configurations, e.redisQueue()...)
	configurations = append(configurations, e.redisSocketio()...)
	configurations = append(configurations, e.mariadb()...)
	configurations = append(configurations, e.erpnext()...)
	return configurations
}

func (e *ErpNext) redisCache() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/redis",
		Tag:        "4.0.11-alpine",
	}
	port := deployer.Port{
		Port:     13000,
		Protocol: "TCP",
		Name:     "redis",
	}
	return []world.Configuration{
		world.NewConfiguraionBuilder().WithApplier(
			&deployer.ConfigMapApplier{
				Context:   e.Context,
				Namespace: "erpnext",
				Name:      "redis-cache",
				ConfigEntryList: deployer.ConfigEntryList{
					deployer.ConfigEntry{
						Key:   "redis.conf",
						Value: redisCacheConf,
					},
				},
			},
		),
		&deployer.DeploymentDeployer{
			Context:   e.Context,
			Namespace: "erpnext",
			Name:      "redis-cache",
			Strategy: k8s.DeploymentStrategy{
				Type: "Recreate",
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name:    "redis",
					Command: []k8s.Command{"redis-server", "/etc/conf.d/redis.conf"},
					Image:   image,
					Requirement: &build.Redis{
						Image: image,
					},
					Ports: []deployer.Port{
						port,
					},
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "500m",
							Memory: "100Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "10m",
							Memory: "10Mi",
						},
					},
					LivenessProbe: k8s.Probe{
						TcpSocket: k8s.TcpSocket{
							Port: port.Port,
						},
						InitialDelaySeconds: 60,
						SuccessThreshold:    1,
						FailureThreshold:    5,
						TimeoutSeconds:      5,
						PeriodSeconds:       10,
					},
					ReadinessProbe: k8s.Probe{
						TcpSocket: k8s.TcpSocket{
							Port: port.Port,
						},
						InitialDelaySeconds: 15,
						TimeoutSeconds:      5,
						PeriodSeconds:       10,
					},
					Mounts: []k8s.ContainerMount{
						{
							Name: "config",
							Path: "/etc/conf.d",
						},
					},
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name: "config",
					ConfigMap: k8s.PodVolumeConfigMap{
						Name: "redis-cache",
						Items: []k8s.PodConfigMapItem{
							{
								Key:  "redis.conf",
								Path: "redis.conf",
							},
						},
					},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   e.Context,
			Namespace: "erpnext",
			Name:      "redis-cache",
			Ports:     []deployer.Port{port},
		},
	}
}

func (e *ErpNext) redisQueue() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/redis",
		Tag:        "4.0.11-alpine",
	}
	port := deployer.Port{
		Port:     11000,
		Protocol: "TCP",
		Name:     "redis",
	}
	return []world.Configuration{
		world.NewConfiguraionBuilder().WithApplier(
			&deployer.ConfigMapApplier{
				Context:   e.Context,
				Namespace: "erpnext",
				Name:      "redis-queue",
				ConfigEntryList: deployer.ConfigEntryList{
					deployer.ConfigEntry{
						Key:   "redis.conf",
						Value: redisQueueConf,
					},
				},
			},
		),
		&deployer.DeploymentDeployer{
			Context:   e.Context,
			Namespace: "erpnext",
			Name:      "redis-queue",
			Strategy: k8s.DeploymentStrategy{
				Type: "Recreate",
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name:    "redis",
					Command: []k8s.Command{"redis-server", "/etc/conf.d/redis.conf"},
					Image:   image,
					Requirement: &build.Redis{
						Image: image,
					},
					Ports: []deployer.Port{
						port,
					},
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "500m",
							Memory: "100Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "10m",
							Memory: "10Mi",
						},
					},
					LivenessProbe: k8s.Probe{
						TcpSocket: k8s.TcpSocket{
							Port: port.Port,
						},
						InitialDelaySeconds: 60,
						SuccessThreshold:    1,
						FailureThreshold:    5,
						TimeoutSeconds:      5,
						PeriodSeconds:       10,
					},
					ReadinessProbe: k8s.Probe{
						TcpSocket: k8s.TcpSocket{
							Port: port.Port,
						},
						InitialDelaySeconds: 15,
						TimeoutSeconds:      5,
						PeriodSeconds:       10,
					},
					Mounts: []k8s.ContainerMount{
						{
							Name: "config",
							Path: "/etc/conf.d",
						},
					},
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name: "config",
					ConfigMap: k8s.PodVolumeConfigMap{
						Name: "redis-queue",
						Items: []k8s.PodConfigMapItem{
							{
								Key:  "redis.conf",
								Path: "redis.conf",
							},
						},
					},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   e.Context,
			Namespace: "erpnext",
			Name:      "redis-queue",
			Ports:     []deployer.Port{port},
		},
	}
}

func (e *ErpNext) redisSocketio() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/redis",
		Tag:        "4.0.11-alpine",
	}
	port := deployer.Port{
		Port:     12000,
		Protocol: "TCP",
		Name:     "redis",
	}
	return []world.Configuration{
		world.NewConfiguraionBuilder().WithApplier(
			&deployer.ConfigMapApplier{
				Context:   e.Context,
				Namespace: "erpnext",
				Name:      "redis-socketio",
				ConfigEntryList: deployer.ConfigEntryList{
					deployer.ConfigEntry{
						Key:   "redis.conf",
						Value: redisSocketioConf,
					},
				},
			},
		),
		&deployer.DeploymentDeployer{
			Context:   e.Context,
			Namespace: "erpnext",
			Name:      "redis-socketio",
			Strategy: k8s.DeploymentStrategy{
				Type: "Recreate",
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name:    "redis",
					Command: []k8s.Command{"redis-server", "/etc/conf.d/redis.conf"},
					Image:   image,
					Requirement: &build.Redis{
						Image: image,
					},
					Ports: []deployer.Port{
						port,
					},
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "500m",
							Memory: "100Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "10m",
							Memory: "10Mi",
						},
					},
					LivenessProbe: k8s.Probe{
						TcpSocket: k8s.TcpSocket{
							Port: port.Port,
						},
						InitialDelaySeconds: 60,
						SuccessThreshold:    1,
						FailureThreshold:    5,
						TimeoutSeconds:      5,
						PeriodSeconds:       10,
					},
					ReadinessProbe: k8s.Probe{
						TcpSocket: k8s.TcpSocket{
							Port: port.Port,
						},
						InitialDelaySeconds: 15,
						TimeoutSeconds:      5,
						PeriodSeconds:       10,
					},
					Mounts: []k8s.ContainerMount{
						{
							Name: "config",
							Path: "/etc/conf.d",
						},
					},
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name: "config",
					ConfigMap: k8s.PodVolumeConfigMap{
						Name: "redis-socketio",
						Items: []k8s.PodConfigMapItem{
							{
								Key:  "redis.conf",
								Path: "redis.conf",
							},
						},
					},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   e.Context,
			Namespace: "erpnext",
			Name:      "redis-socketio",
			Ports:     []deployer.Port{port},
		},
	}
}

func (e *ErpNext) mariadb() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/mariadb",
		Tag:        "10.3.9",
	}
	port := deployer.Port{
		Port:     3306,
		Protocol: "TCP",
		Name:     "mariadb",
	}
	return []world.Configuration{
		world.NewConfiguraionBuilder().WithApplier(
			&deployer.ConfigMapApplier{
				Context:   e.Context,
				Namespace: "erpnext",
				Name:      "mariadb",
				ConfigEntryList: deployer.ConfigEntryList{
					deployer.ConfigEntry{
						Key:   "my.cnf",
						Value: mariadbMyCnf,
					},
					deployer.ConfigEntry{
						Key:   "mysql.cnf",
						Value: mariadbMysqlCnf,
					},
					deployer.ConfigEntry{
						Key:   "mysqld_safe_syslog.cnf",
						Value: mariadbMysqldSafeSyslogCnf,
					},
					deployer.ConfigEntry{
						Key:   "mysqldump.cnf",
						Value: mariadbMysqldumpCnf,
					},
				},
			},
		),
		&deployer.SecretDeployer{
			Context:   e.Context,
			Namespace: "erpnext",
			Name:      "mariadb",
			Secrets: deployer.Secrets{
				"root-password": e.DatabaseRootPassword,
				"db-name":       e.DatabaseName,
				"db-password":   e.DatabasePassword,
			},
		},
		&deployer.DeploymentDeployer{
			Context:   e.Context,
			Namespace: "erpnext",
			Name:      "mariadb",
			Strategy: k8s.DeploymentStrategy{
				Type: "Recreate",
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name:  "mariadb",
					Image: image,
					Env: []k8s.Env{
						{
							Name:  "MYSQL_ROOT_HOST",
							Value: "%",
						},
						{
							Name: "MYSQL_ROOT_PASSWORD",
							ValueFrom: k8s.ValueFrom{
								SecretKeyRef: k8s.SecretKeyRef{
									Key:  "root-password",
									Name: "mariadb",
								},
							},
						},
						{
							Name: "MYSQL_USER",
							ValueFrom: k8s.ValueFrom{
								SecretKeyRef: k8s.SecretKeyRef{
									Key:  "db-name",
									Name: "mariadb",
								},
							},
						},
						{
							Name: "MYSQL_DATABASE",
							ValueFrom: k8s.ValueFrom{
								SecretKeyRef: k8s.SecretKeyRef{
									Key:  "db-name",
									Name: "mariadb",
								},
							},
						},
						{
							Name: "MYSQL_PASSWORD",
							ValueFrom: k8s.ValueFrom{
								SecretKeyRef: k8s.SecretKeyRef{
									Key:  "db-password",
									Name: "mariadb",
								},
							},
						},
					},
					Requirement: &build.Mariadb{
						Image: image,
					},
					Ports: []deployer.Port{
						port,
					},
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "1000m",
							Memory: "450Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "100m",
							Memory: "150Mi",
						},
					},
					LivenessProbe: k8s.Probe{
						TcpSocket: k8s.TcpSocket{
							Port: port.Port,
						},
						InitialDelaySeconds: 60,
						SuccessThreshold:    1,
						FailureThreshold:    5,
						TimeoutSeconds:      5,
						PeriodSeconds:       10,
					},
					ReadinessProbe: k8s.Probe{
						TcpSocket: k8s.TcpSocket{
							Port: port.Port,
						},
						InitialDelaySeconds: 15,
						TimeoutSeconds:      5,
						PeriodSeconds:       10,
					},
					Mounts: []k8s.ContainerMount{
						{
							Name: "data",
							Path: "/var/lib/mysql",
						},
						{
							Name: "config",
							Path: "/etc/mysql/conf.d",
						},
					},
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name: "data",
					Nfs: k8s.PodVolumeNfs{
						Path:   "/data/erpnext-mariadb",
						Server: k8s.PodNfsServer(e.NfsServer),
					},
				},
				{
					Name: "config",
					ConfigMap: k8s.PodVolumeConfigMap{
						Name: "mariadb",
						Items: []k8s.PodConfigMapItem{
							{
								Key:  "my.cnf",
								Path: "my.cnf",
							},
							{
								Key:  "mysql.cnf",
								Path: "mysql.cnf",
							},
							{
								Key:  "mysqld_safe_syslog.cnf",
								Path: "mysqld_safe_syslog.cnf",
							},
							{
								Key:  "mysqldump.cnf",
								Path: "mysqldump.cnf",
							},
						},
					},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   e.Context,
			Namespace: "erpnext",
			Name:      "mariadb",
			Ports:     []deployer.Port{port},
		},
	}
}

func (e *ErpNext) erpnext() []world.Configuration {
	erpnextImage := docker.Image{
		Repository: "bborbe/erpnext",
		Tag:        "1.0.1",
	}
	webserverPort := deployer.Port{
		Port:     8000,
		Protocol: "TCP",
		Name:     "webserver",
	}
	socketioPort := deployer.Port{
		Port:     9000,
		Protocol: "TCP",
		Name:     "socketio",
	}
	fileWatcherPort := deployer.Port{
		Port:     6787,
		Protocol: "TCP",
		Name:     "filewatcher",
	}
	ports := []deployer.Port{
		webserverPort,
		socketioPort,
		fileWatcherPort,
	}
	return []world.Configuration{
		&deployer.SecretDeployer{
			Context:   e.Context,
			Namespace: "erpnext",
			Name:      "erpnext",
			Secrets: deployer.Secrets{
				"db-name":        e.DatabaseName,
				"db-password":    e.DatabasePassword,
				"admin-password": e.AdminPassword,
			},
		},
		&deployer.DeploymentDeployer{
			Context:   e.Context,
			Namespace: "erpnext",
			Name:      "erpnext",
			Strategy: k8s.DeploymentStrategy{
				Type: "Recreate",
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name:  "erpnext",
					Image: erpnextImage,
					Requirement: &build.Erpnext{
						Image: erpnextImage,
					},
					Env: []k8s.Env{
						{
							Name:  "DB_HOST",
							Value: "mariadb",
						},
						{
							Name: "DB_NAME",
							ValueFrom: k8s.ValueFrom{
								SecretKeyRef: k8s.SecretKeyRef{
									Key:  "db-name",
									Name: "erpnext",
								},
							},
						},
						{
							Name: "DB_PASSWORD",
							ValueFrom: k8s.ValueFrom{
								SecretKeyRef: k8s.SecretKeyRef{
									Key:  "db-password",
									Name: "erpnext",
								},
							},
						},
						{
							Name: "ADMIN_PASSWORD",
							ValueFrom: k8s.ValueFrom{
								SecretKeyRef: k8s.SecretKeyRef{
									Key:  "admin-password",
									Name: "erpnext",
								},
							},
						},
					},
					Ports: ports,
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "1000m",
							Memory: "2000Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "100m",
							Memory: "1000Mi",
						},
					},
					LivenessProbe: k8s.Probe{
						HttpGet: k8s.HttpGet{
							Path:   "/",
							Port:   webserverPort.Port,
							Scheme: "HTTP",
						},
						InitialDelaySeconds: 120,
						SuccessThreshold:    1,
						FailureThreshold:    5,
						TimeoutSeconds:      5,
					},
					ReadinessProbe: k8s.Probe{
						HttpGet: k8s.HttpGet{
							Path:   "/",
							Port:   webserverPort.Port,
							Scheme: "HTTP",
						},
						InitialDelaySeconds: 60,
						TimeoutSeconds:      5,
					},
					Mounts: []k8s.ContainerMount{
						{
							Name: "erpnext-frappe-public",
							Path: "/home/frappe/branch-repo/sites/site1.local/public",
						},
						{
							Name: "erpnext-frappe-private",
							Path: "/home/frappe/branch-repo/sites/site1.local/private",
						},
					},
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name: "erpnext-frappe-public",
					Nfs: k8s.PodVolumeNfs{
						Path:   "/data/erpnext-frappe/public",
						Server: k8s.PodNfsServer(e.NfsServer),
					},
				},
				{
					Name: "erpnext-frappe-private",
					Nfs: k8s.PodVolumeNfs{
						Path:   "/data/erpnext-frappe/private",
						Server: k8s.PodNfsServer(e.NfsServer),
					},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   e.Context,
			Namespace: "erpnext",
			Name:      "erpnext",
			Ports:     ports,
		},
		world.NewConfiguraionBuilder().WithApplier(
			&k8s.IngressApplier{
				Context: e.Context,
				Ingress: k8s.Ingress{
					ApiVersion: "extensions/v1beta1",
					Kind:       "Ingress",
					Metadata: k8s.Metadata{
						Namespace: "erpnext",
						Name:      "erpnext",
						Annotations: k8s.Annotations{
							"kubernetes.io/ingress.class": "traefik",
						},
					},
					Spec: k8s.IngressSpec{
						Rules: []k8s.IngressRule{
							{
								Host: e.Domain,
								Http: k8s.IngressHttp{
									Paths: []k8s.IngressPath{
										{
											Path: "/",
											Backends: k8s.IngressBackend{
												ServiceName: k8s.IngressBackendServiceName("erpnext"),
												ServicePort: k8s.IngressBackendServicePort(webserverPort.Name),
											},
										},
									},
								},
							},
							{
								Host: e.Domain,
								Http: k8s.IngressHttp{
									Paths: []k8s.IngressPath{
										{
											Path: "/socket.io/",
											Backends: k8s.IngressBackend{
												ServiceName: k8s.IngressBackendServiceName("erpnext"),
												ServicePort: k8s.IngressBackendServicePort(socketioPort.Name),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		),
	}
}

const mariadbMyCnf = `
[client]

[mysqld]
bind-address = 0.0.0.0
!includedir /etc/mysql/mariadb.conf.d/

[mysqld]
character-set-client-handshake = FALSE
character-set-server = utf8mb4
collation-server = utf8mb4_unicode_ci

[mysql]
default-character-set = utf8mb4
`

const mariadbMysqlCnf = `
[mysql]
`

const mariadbMysqldSafeSyslogCnf = `
[mysqld_safe]
skip_log_error
syslog
`

const mariadbMysqldumpCnf = `
[mysqldump]
quick
quote-names
max_allowed_packet	= 16M
`

const redisQueueConf = `
dbfilename redis_queue.rdb
bind 0.0.0.0
port 11000
`

const redisSocketioConf = `
dbfilename redis_socketio.rdb
bind 0.0.0.0
port 12000
`

const redisCacheConf = `
dbfilename redis_cache.rdb
bind 0.0.0.0
port 13000
maxmemory 292mb
maxmemory-policy allkeys-lru
appendonly no

save ""
`
