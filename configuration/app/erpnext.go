package app

import (
	"context"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type ErpNext struct {
	Cluster           cluster.Cluster
	Domain            k8s.IngressHost
	MysqlRootPassword deployer.SecretValue
}

func (e *ErpNext) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		e.Cluster,
		e.Domain,
		e.MysqlRootPassword,
	)
}

func (e *ErpNext) Applier() (world.Applier, error) {
	return nil, nil
}

func (e *ErpNext) Children() []world.Configuration {
	configurations := []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   e.Cluster.Context,
			Namespace: "erpnext",
		},
	}
	configurations = append(configurations, e.redisCache()...)
	configurations = append(configurations, e.redisQueue()...)
	configurations = append(configurations, e.redisSocketio()...)
	configurations = append(configurations, e.mariadb()...)
	configurations = append(configurations, e.frappe()...)
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
		&deployer.ConfigMapDeployer{
			Context:   e.Cluster.Context,
			Namespace: "erpnext",
			Name:      "redis-cache",
			ConfigMapData: k8s.ConfigMapData{
				"redis.conf": redisCacheConf,
			},
		},
		&deployer.DeploymentDeployer{
			Context:   e.Cluster.Context,
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
			Context:   e.Cluster.Context,
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
		&deployer.ConfigMapDeployer{
			Context:   e.Cluster.Context,
			Namespace: "erpnext",
			Name:      "redis-queue",
			ConfigMapData: k8s.ConfigMapData{
				"redis.conf": redisQueueConf,
			},
		},
		&deployer.DeploymentDeployer{
			Context:   e.Cluster.Context,
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
			Context:   e.Cluster.Context,
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
		&deployer.ConfigMapDeployer{
			Context:   e.Cluster.Context,
			Namespace: "erpnext",
			Name:      "redis-socketio",
			ConfigMapData: k8s.ConfigMapData{
				"redis.conf": redisSocketioConf,
			},
		},
		&deployer.DeploymentDeployer{
			Context:   e.Cluster.Context,
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
			Context:   e.Cluster.Context,
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
		&deployer.ConfigMapDeployer{
			Context:   e.Cluster.Context,
			Namespace: "erpnext",
			Name:      "mariadb",
			ConfigMapData: k8s.ConfigMapData{
				"my.cnf":                 mariadbMyCnf,
				"mysql.cnf":              mariadbMysqlCnf,
				"mysqld_safe_syslog.cnf": mariadbMysqldSafeSyslogCnf,
				"mysqldump.cnf":          mariadbMysqldumpCnf,
			},
		},
		&deployer.SecretDeployer{
			Context:   e.Cluster.Context,
			Namespace: "erpnext",
			Name:      "mariadb",
			Secrets: deployer.Secrets{
				"mysql-root-password": e.MysqlRootPassword,
			},
		},
		&deployer.DeploymentDeployer{
			Context:   e.Cluster.Context,
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
							Name:  "MYSQL_USER",
							Value: "root",
						},
						{
							Name:  "MYSQL_DATABASE",
							Value: "erpnext",
						},
						{
							Name: "MYSQL_ROOT_PASSWORD",
							ValueFrom: k8s.ValueFrom{
								SecretKeyRef: k8s.SecretKeyRef{
									Key:  "mysql-root-password",
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
						Server: k8s.PodNfsServer(e.Cluster.NfsServer),
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
			Context:   e.Cluster.Context,
			Namespace: "erpnext",
			Name:      "mariadb",
			Ports:     []deployer.Port{port},
		},
	}
}

func (e *ErpNext) frappe() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/frappe",
		Tag:        "master",
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
		&deployer.DeploymentDeployer{
			Context:   e.Cluster.Context,
			Namespace: "erpnext",
			Name:      "frappe",
			Strategy: k8s.DeploymentStrategy{
				Type: "Recreate",
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name:    "frappe",
					Image:   image,
					Command: []k8s.Command{"bench", "start"},
					Requirement: &build.Frappe{
						Image: image,
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
						InitialDelaySeconds: 60,
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
						InitialDelaySeconds: 15,
						TimeoutSeconds:      5,
					},
					Mounts: []k8s.ContainerMount{
						{
							Name: "data",
							Path: "/home/frappe/frappe-bench",
						},
					},
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name: "data",
					Nfs: k8s.PodVolumeNfs{
						Path:   "/data/erpnext-frappe",
						Server: k8s.PodNfsServer(e.Cluster.NfsServer),
					},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   e.Cluster.Context,
			Namespace: "erpnext",
			Name:      "frappe",
			Ports:     ports,
		},
		&deployer.IngressDeployer{
			Context:   e.Cluster.Context,
			Namespace: "erpnext",
			Name:      "frappe",
			Port:      webserverPort.Name,
			Domains:   k8s.IngressHosts{e.Domain},
		},
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
