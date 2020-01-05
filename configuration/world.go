// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package configuration

import (
	"context"

	"github.com/bborbe/world/configuration/server"

	"github.com/bborbe/world/configuration/openvpn"

	"github.com/bborbe/world/configuration/app"
	"github.com/bborbe/world/configuration/backup"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/network"
	"github.com/bborbe/world/configuration/service"
	"github.com/bborbe/world/pkg/dns"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/hetzner"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/secret"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type AppName string

type ClusterName string

type World struct {
	App              AppName
	Cluster          ClusterName
	TeamvaultSecrets *secret.Teamvault
	HetznerClient    hetzner.Client
}

func (w *World) Children() []world.Configuration {
	var result []world.Configuration
	for clusterName, configurations := range w.configurations() {
		if clusterName != w.Cluster && w.Cluster != "" {
			continue
		}
		for appName, configuration := range configurations {
			if appName != w.App && w.App != "" {
				continue
			}
			result = append(result, configuration)
		}
	}
	return result
}

func (w *World) Applier() (world.Applier, error) {
	return nil, nil
}

func (w *World) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		w.TeamvaultSecrets,
	)
}

func (w *World) configurations() map[ClusterName]map[AppName]world.Configuration {
	return map[ClusterName]map[AppName]world.Configuration{
		"netcup":    w.netcup(),
		"sun":       w.sun(),
		"fire":      w.fire(),
		"nuke":      w.nuke(),
		"hetzner-1": w.hetzner1(),
		"nova":      w.nova(),
		"rasp":      w.rasp(),
		"co2hz":     w.co2hz(),
		"co2wz":     w.co2wz(),
	}
}

func (w *World) hetzner1() map[AppName]world.Configuration {
	name := "hetzner"
	k8sContext := k8s.Context("hetzner-1")
	apiKey := w.TeamvaultSecrets.Password("kLolmq")
	ip := &hetzner.IP{
		Client: w.HetznerClient,
		ApiKey: apiKey,
		Name:   k8sContext,
	}

	user := ssh.User("bborbe")
	ssh := &ssh.SSH{
		Host: ssh.Host{
			IP:   ip,
			Port: 22,
		},
		User:           user,
		PrivateKeyPath: "/Users/bborbe/.ssh/id_rsa",
	}
	return map[AppName]world.Configuration{
		"cluster": &cluster.Hetzner{
			Context:    k8sContext,
			ApiKey:     apiKey,
			IP:         ip,
			ServerType: "cx11",
			User:       user,
			SSH:        ssh,
		},
		"openvpn-server": &openvpn.Server{
			SSH:         ssh,
			ServerName:  openvpn.ServerName(name),
			ServerIPNet: network.HetznerVPN.IPNet,
			Routes:      openvpn.Routes{},
		},
		"cluster-admin": &service.ClusterAdmin{
			Context: k8sContext,
		},
		"calico": &service.Calico{
			Context:   k8sContext,
			ClusterIP: ip,
		},
		"dns": &app.CoreDns{
			Context: k8sContext,
		},
		"kubeless": world.NewConfiguraionBuilder().WithApplier(
			&dns.Server{
				Host:    "ns.rocketsource.de",
				KeyPath: "/Users/bborbe/.dns/home.benjamin-borbe.de.key",
				List: []dns.Entry{
					{
						Host: "kubeless.hetzner-1.benjamin-borbe.de",
						IP:   ip,
					},
				},
			},
		),
		"hostpath": &app.HostPathProvisioner{
			Context:             k8sContext,
			HostPath:            "/data",
			DefaultStorageClass: true,
		},
		"traefik": &app.Traefik{
			Context: k8sContext,
			Domains: k8s.IngressHosts{
				"traefik.hetzner-1.benjamin-borbe.de",
			},
		},
		"ip": &app.Ip{
			Context: k8sContext,
			IP:      ip,
			Tag:     "1.1.0",
			Domain:  "ip.hetzner-1.benjamin-borbe.de",
		},
		"openfaas": &app.OpenFaas{
			Context: k8sContext,
			IP:      ip,
			Domain:  "openfaas.hetzner-1.benjamin-borbe.de",
		},
	}
}

func (w *World) nova() map[AppName]world.Configuration {
	server := server.Nova
	return map[AppName]world.Configuration{
		"cluster": &cluster.Nova{
			IP:   server.IP,
			Host: "nova.hm.benjamin-borbe.de",
		},
		"openvpn-client": &openvpn.LocalClient{
			ClientName:    openvpn.ClientName(server.Name),
			ServerName:    "hetzner",
			ServerAddress: "hetzner-1.benjamin-borbe.de",
		},
	}
}
func (w *World) fire() map[AppName]world.Configuration {
	server := server.Fire
	ssh := &ssh.SSH{
		Host: ssh.Host{
			IP:   server.IP,
			Port: 22,
		},
		User:           "bborbe",
		PrivateKeyPath: "/Users/bborbe/.ssh/id_rsa",
	}
	return map[AppName]world.Configuration{
		"cluster": &cluster.Fire{
			SSH:       ssh,
			Context:   k8s.Context(server.Name),
			ClusterIP: server.IP,
		},
		"cluster-admin": &service.ClusterAdmin{
			Context: k8s.Context(server.Name),
		},
		"calico": &service.Calico{
			Context:   k8s.Context(server.Name),
			ClusterIP: server.IP,
		},
		"dns": &app.CoreDns{
			Context: k8s.Context(server.Name),
		},
		"openvpn-client": &openvpn.RemoteClient{
			SSH:           ssh,
			ClientName:    openvpn.ClientName(server.Name),
			ServerName:    "hetzner",
			ServerAddress: "hetzner-1.benjamin-borbe.de",
		},
		"traefik": &app.Traefik{
			Context: k8s.Context(server.Name),
			Domains: k8s.IngressHosts{
				"traefik.fire.hm.benjamin-borbe.de",
			},
		},
		"backup": &app.BackupClient{
			Context: k8s.Context(server.Name),
			Domains: k8s.IngressHosts{
				"backup.fire.hm.benjamin-borbe.de",
			},
			BackupSshKey: w.TeamvaultSecrets.File("8q1bJw"),
			BackupTargets: app.BackupTargets{
				backup.Sun,
			},
		},
	}
}

func (w *World) nuke() map[AppName]world.Configuration {
	server := server.Nuke
	ssh := &ssh.SSH{
		Host: ssh.Host{
			IP:   server.IP,
			Port: 22,
		},
		User:           "bborbe",
		PrivateKeyPath: "/Users/bborbe/.ssh/id_rsa",
	}
	return map[AppName]world.Configuration{
		"cluster": &cluster.Nuke{
			SSH:       ssh,
			Context:   k8s.Context(server.Name),
			ClusterIP: server.IP,
		},
		"cluster-admin": &service.ClusterAdmin{
			Context: k8s.Context(server.Name),
		},
		"calico": &service.Calico{
			Context:   k8s.Context(server.Name),
			ClusterIP: server.IP,
		},
		"dns": &app.CoreDns{
			Context: k8s.Context(server.Name),
		},
		"traefik": &app.Traefik{
			Context: k8s.Context(server.Name),
			Domains: k8s.IngressHosts{
				"traefik.nuke.hm.benjamin-borbe.de",
			},
		},
		"openvpn-client": &openvpn.RemoteClient{
			SSH:           ssh,
			ClientName:    openvpn.ClientName(server.Name),
			ServerName:    "hetzner",
			ServerAddress: "hetzner-1.benjamin-borbe.de",
		},
		"backup": &app.BackupClient{
			Context: k8s.Context(server.Name),
			Domains: k8s.IngressHosts{
				"backup.nuke.hm.benjamin-borbe.de",
			},
			BackupSshKey: w.TeamvaultSecrets.File("8q1bJw"),
			BackupTargets: app.BackupTargets{
				backup.Fire,
			},
		},
	}
}

func (w *World) sun() map[AppName]world.Configuration {
	server := server.Sun
	ssh := &ssh.SSH{
		Host: ssh.Host{
			IP:   server.IP,
			Port: 22,
		},
		User:           "bborbe",
		PrivateKeyPath: "/Users/bborbe/.ssh/id_rsa",
	}
	return map[AppName]world.Configuration{
		"cluster": &cluster.Sun{
			SSH:       ssh,
			Context:   k8s.Context(server.Name),
			ClusterIP: server.IP,
		},
		"cluster-admin": &service.ClusterAdmin{
			Context: k8s.Context(server.Name),
		},
		"calico": &service.Calico{
			Context:   k8s.Context(server.Name),
			ClusterIP: server.IP,
		},
		"dns": &app.CoreDns{
			Context: k8s.Context(server.Name),
		},
		"openvpn-client": &openvpn.RemoteClient{
			SSH:           ssh,
			ClientName:    openvpn.ClientName(server.Name),
			ServerName:    "hetzner",
			ServerAddress: "hetzner-1.benjamin-borbe.de",
		},
		"minecraft": &app.Minecraft{
			Context: k8s.Context(server.Name),
		},
		"monitoring": &app.Monitoring{
			Context:         k8s.Context(server.Name),
			GitSyncPassword: w.TeamvaultSecrets.Password("YLb4wV"),
			SmtpPassword:    w.TeamvaultSecrets.Password("QL3VQO"),
			Configs: []app.MonitoringConfig{
				{
					Name:       "nc",
					Subject:    "Monitoring Result: Netcup",
					GitRepoUrl: "https://bborbereadonly@bitbucket.org/bborbe/monitoring_nc.git",
				},
				{
					Name:       "pn-intern",
					Subject:    "Monitoring Result: PN-Intern",
					GitRepoUrl: "https://bborbereadonly@bitbucket.org/bborbe/monitoring_pn_intern.git",
				},
				{
					Name:       "hm",
					Subject:    "Monitoring Result: HM",
					GitRepoUrl: "https://bborbereadonly@bitbucket.org/bborbe/monitoring_hm.git",
				},
			},
		},
		"traefik": &app.Traefik{
			Context: k8s.Context(server.Name),
			Domains: k8s.IngressHosts{
				"traefik.sun.pn.benjamin-borbe.de",
			},
		},
		"backup": &app.BackupClient{
			Context: k8s.Context(server.Name),
			Domains: k8s.IngressHosts{
				"backup.sun.pn.benjamin-borbe.de",
			},
			BackupSshKey: w.TeamvaultSecrets.File("8q1bJw"),
			BackupTargets: app.BackupTargets{
				backup.Netcup,
				backup.Sun,
				backup.Rasp,
				backup.Co2hz,
				backup.Co2wz,
				backup.Fire,
				backup.Nuke,
				backup.Nova,
				backup.Star,
			},
		},
	}
}

func (w *World) netcup() map[AppName]world.Configuration {
	server := server.Netcup
	return map[AppName]world.Configuration{
		"cluster": &cluster.Netcup{
			Context:     k8s.Context(server.Name),
			IP:          server.IP,
			DisableCNI:  true,
			DisableRBAC: true,
		},
		"cluster-admin": &service.ClusterAdmin{
			Context: k8s.Context(server.Name),
		},
		"calico": &service.Calico{
			Context:   k8s.Context(server.Name),
			ClusterIP: server.IP,
		},
		"debug": &app.Debug{
			Context: k8s.Context(server.Name),
			Domain:  "debug.benjamin-borbe.de",
			Requirements: []world.Configuration{
				world.NewConfiguraionBuilder().WithApplier(
					&dns.Server{
						Host:    "ns.rocketsource.de",
						KeyPath: "/Users/bborbe/.dns/home.benjamin-borbe.de.key",
						List: []dns.Entry{
							{
								Host: network.DebugHostname,
								IP:   server.IP,
							},
						},
					},
				),
			},
		},
		"mqtt-kafka-connector-co2mon": &app.MqttKafkaConnector{
			Context:      k8s.Context(server.Name),
			MqttBroker:   "tcp://rasp.hm.benjamin-borbe.de:1883",
			MqttUser:     w.TeamvaultSecrets.Username("9qNx3O"),
			MqttPassword: w.TeamvaultSecrets.Password("9qNx3O"),
			MqttTopic:    "co2mon",
			KafkaBrokers: []string{"kafka-cp-kafka-headless.kafka.svc.cluster.local:9092"},
			KafkaTopic:   "co2mon",
		},
		"metabase": &app.Metabase{
			Context:          k8s.Context(server.Name),
			Domain:           "metabase.benjamin-borbe.de",
			DatabasePassword: w.TeamvaultSecrets.Password("dwkWAw"),
			Requirements: []world.Configuration{
				world.NewConfiguraionBuilder().WithApplier(
					&dns.Server{
						Host:    "ns.rocketsource.de",
						KeyPath: "/Users/bborbe/.dns/home.benjamin-borbe.de.key",
						List: []dns.Entry{
							{
								Host: network.MetabaseHostname,
								IP:   server.IP,
							},
						},
					},
				),
			},
		},
		"grafana": &app.Grafana{
			Context:      k8s.Context(server.Name),
			Domain:       k8s.IngressHost(network.GrafanaHostname.String()),
			LdapUsername: w.TeamvaultSecrets.Username("MOPMLG"),
			LdapPassword: w.TeamvaultSecrets.Password("MOPMLG"),
			Requirements: []world.Configuration{
				world.NewConfiguraionBuilder().WithApplier(
					&dns.Server{
						Host:    "ns.rocketsource.de",
						KeyPath: "/Users/bborbe/.dns/home.benjamin-borbe.de.key",
						List: []dns.Entry{
							{
								Host: network.GrafanaHostname,
								IP:   server.IP,
							},
						},
					},
				),
			},
		},
		"hostpath": &app.HostPathProvisioner{
			Context:             k8s.Context(server.Name),
			HostPath:            "/data/hostpath-provisioner",
			DefaultStorageClass: true,
		},
		"kafka": &app.Kafka{
			AccessMode:        "ReadWriteMany",
			Context:           k8s.Context(server.Name),
			DisableConnect:    true,
			DisableRest:       true,
			KafkaReplicas:     1,
			KafkaStorage:      "5Gi",
			StorageClass:      "hostpath",
			ZookeeperReplicas: 1,
			ZookeeperStorage:  "5Gi",
			Version:           "5.3.1", // https://hub.docker.com/r/confluentinc/cp-kafka/tags
		},
		"kafka-status": &app.KafkaStatus{
			Context:  k8s.Context(server.Name),
			Replicas: 1,
			Domain:   "kafka-status.benjamin-borbe.de",
			Requirements: []world.Configuration{
				world.NewConfiguraionBuilder().WithApplier(
					&dns.Server{
						Host:    "ns.rocketsource.de",
						KeyPath: "/Users/bborbe/.dns/home.benjamin-borbe.de.key",
						List: []dns.Entry{
							{
								Host: network.KafkaStatus,
								IP:   server.IP,
							},
						},
					},
				),
			},
		},
		"kafka-latest-versions": &app.KafkaLatestVersions{
			Context:      k8s.Context(server.Name),
			Replicas:     2,
			AccessMode:   "ReadWriteMany",
			StorageClass: "hostpath",
			Domain:       "versions.benjamin-borbe.de",
			Requirements: []world.Configuration{
				world.NewConfiguraionBuilder().WithApplier(
					&dns.Server{
						Host:    "ns.rocketsource.de",
						KeyPath: "/Users/bborbe/.dns/home.benjamin-borbe.de.key",
						List: []dns.Entry{
							{
								Host: network.VersionsHostname,
								IP:   server.IP,
							},
						},
					},
				),
			},
		},
		"kafka-update-available": &app.KafkaUpdateAvailable{
			Context:      k8s.Context(server.Name),
			Replicas:     2,
			AccessMode:   "ReadWriteMany",
			StorageClass: "hostpath",
			Domain:       "updates.benjamin-borbe.de",
			Requirements: []world.Configuration{
				world.NewConfiguraionBuilder().WithApplier(
					&dns.Server{
						Host:    "ns.rocketsource.de",
						KeyPath: "/Users/bborbe/.dns/home.benjamin-borbe.de.key",
						List: []dns.Entry{
							{
								Host: network.UpdatesHostname,
								IP:   server.IP,
							},
						},
					},
				),
			},
		},
		"kafka-k8s-version-collector": &app.KafkaK8sVersionCollector{
			Context: k8s.Context(server.Name),
		},
		"kafka-dockerhub-version-collector": &app.KafkaDockerhubVersionCollector{
			Context: k8s.Context(server.Name),
			Repositories: []docker.Repository{
				"confluentinc/cp-kafka-connect",
				"confluentinc/cp-kafka-rest",
				"confluentinc/cp-kafka-rest",
				"confluentinc/cp-kafka",
				"confluentinc/cp-ksql-server",
				"confluentinc/cp-ksql-server",
				"confluentinc/cp-schema-registry",
				"confluentinc/cp-zookeeper",
				"coredns/coredns",
				"grafana/grafana",
				"jrelva/nginx-autoindex",
				"library/alpine",
				"library/golang",
				"library/mariadb",
				"library/postgres",
				"library/redis",
				"library/traefik",
				"library/ubuntu",
				"metabase/metabase",
			},
		},
		"kafka-regex-version-collector": &app.KafkaRegexVersionCollector{
			Context:     k8s.Context(server.Name),
			Application: "Golang",
			Url:         "https://golang.org/dl/",
			Regex:       `https://dl.google.com/go/go([\d\.]+)\.src\.tar\.gz`,
		},
		"kafka-atlassian-version-collector": &app.KafkaAtlassianVersionCollector{
			Context: k8s.Context(server.Name),
		},
		"kafka-installed-version-collector": &app.KafkaInstalledVersionCollector{
			Context: k8s.Context(server.Name),
			Apps: []struct {
				Name  string
				Regex string
				Url   string
			}{
				{
					Name:  "Confluence",
					Regex: `<meta\s+name="ajs-version-number"\s+content="([^"]+)">`,
					Url:   "https://confluence.benjamin-borbe.de",
				},
				{
					Name:  "Jira",
					Regex: `<meta\s+name="ajs-version-number"\s+content="([^"]+)">`,
					Url:   "https://jira.benjamin-borbe.de",
				},
			},
		},
		"kafka-sample": &app.KafkaSample{
			Context: k8s.Context(server.Name),
			Domain:  "kafka-sample.benjamin-borbe.de",
			Requirements: []world.Configuration{
				world.NewConfiguraionBuilder().WithApplier(
					&dns.Server{
						Host:    "ns.rocketsource.de",
						KeyPath: "/Users/bborbe/.dns/home.benjamin-borbe.de.key",
						List: []dns.Entry{
							{
								Host: network.KafkaSampleHostname,
								IP:   server.IP,
							},
						},
					},
				),
			},
		},
		"dns": &app.CoreDns{
			Context: k8s.Context(server.Name),
		},
		"traefik": &app.Traefik{
			Context: k8s.Context(server.Name),
			Domains: k8s.IngressHosts{
				"traefik.benjamin-borbe.de",
			},
			SSL: true,
		},
		"prometheus": &app.Prometheus{
			Context:            k8s.Context(server.Name),
			PrometheusDomain:   "prometheus.benjamin-borbe.de",
			AlertmanagerDomain: "prometheus-alertmanager.benjamin-borbe.de",
			Secret:             w.TeamvaultSecrets.Password("aqMr6w"),
			LdapUsername:       w.TeamvaultSecrets.Username("MOPMLG"),
			LdapPassword:       w.TeamvaultSecrets.Password("MOPMLG"),
		},
		"ldap": &app.Ldap{
			Context:      k8s.Context(server.Name),
			Tag:          "1.3.0",
			LdapPassword: w.TeamvaultSecrets.Password("MOPMLG"),
		},
		"teamvault": &app.Teamvault{
			Context:          k8s.Context(server.Name),
			Domain:           "teamvault.benjamin-borbe.de",
			DatabasePassword: w.TeamvaultSecrets.Password("VO0W5w"),
			SmtpUsername:     w.TeamvaultSecrets.Username("3OlNaq"),
			SmtpPassword:     w.TeamvaultSecrets.Password("3OlNaq"),
			LdapPassword:     w.TeamvaultSecrets.Password("MOPMLG"),
			SecretKey:        w.TeamvaultSecrets.Password("NqA68w"),
			FernetKey:        w.TeamvaultSecrets.Password("5wYZ2O"),
			Salt:             w.TeamvaultSecrets.Password("Rwg74w"),
		},
		"monitoring": &app.Monitoring{
			Context:         k8s.Context(server.Name),
			GitSyncPassword: w.TeamvaultSecrets.Password("YLb4wV"),
			SmtpPassword:    w.TeamvaultSecrets.Password("QL3VQO"),
			Configs: []app.MonitoringConfig{
				{
					Name:       "pn",
					Subject:    "Monitoring Result: PN",
					GitRepoUrl: "https://bborbereadonly@bitbucket.org/bborbe/monitoring_pn.git",
				},
				{
					Name:       "hm",
					Subject:    "Monitoring Result: HM",
					GitRepoUrl: "https://bborbereadonly@bitbucket.org/bborbe/monitoring_hm.git",
				},
				{
					Name:       "work",
					Subject:    "Monitoring Result: Work",
					GitRepoUrl: "https://bborbereadonly@bitbucket.org/bborbe/monitoring_work.git",
				},
			},
		},
		"confluence": &app.Confluence{
			Context:          k8s.Context(server.Name),
			Domain:           "confluence.benjamin-borbe.de",
			Version:          "6.15.10",
			DatabasePassword: w.TeamvaultSecrets.Password("3OlaLn"),
			SmtpUsername:     w.TeamvaultSecrets.Username("nOeNjL"),
			SmtpPassword:     w.TeamvaultSecrets.Password("nOeNjL"),
		},
		"jira": &app.Jira{
			Context:          k8s.Context(server.Name),
			Domain:           "jira.benjamin-borbe.de",
			Version:          "7.13.11",
			DatabasePassword: w.TeamvaultSecrets.Password("eOB12w"),
			SmtpUsername:     w.TeamvaultSecrets.Username("MwmE0w"),
			SmtpPassword:     w.TeamvaultSecrets.Password("MwmE0w"),
		},
		"backup": &app.BackupServer{
			Context: k8s.Context(server.Name),
		},
		"poste": &app.Poste{
			Context:      k8s.Context(server.Name),
			PosteVersion: "2.2.0", // https://hub.docker.com/r/analogic/poste.io/tags
			Domains: k8s.IngressHosts{
				"mail.benjamin-borbe.de",
			},
		},
		"maven": &app.Maven{
			Context: k8s.Context(server.Name),
			Domains: k8s.IngressHosts{
				"maven.benjamin-borbe.de",
			},
			MavenRepoVersion: "1.0.0",
		},
		"portfolio": &app.Portfolio{
			Context: k8s.Context(server.Name),
			Domains: k8s.IngressHosts{
				"benjamin-borbe.de",
				"www.benjamin-borbe.de",
				"benjaminborbe.de",
				"www.benjaminborbe.de",
			},
			OverlayServerVersion: "1.0.0",
			GitSyncPassword:      w.TeamvaultSecrets.Password("YLb4wV"),
		},
		"webdav": &app.Webdav{
			Context: k8s.Context(server.Name),
			Domains: k8s.IngressHosts{
				"webdav.benjamin-borbe.de",
			},
			Password: w.TeamvaultSecrets.Password("VOzvAO"),
		},
		"bind": &app.Bind{
			Context: k8s.Context(server.Name),
		},
		"download": &app.Download{
			Context: k8s.Context(server.Name),
			Domains: k8s.IngressHosts{
				"dl.benjamin-borbe.de",
			},
		},
		"mumble": &app.Mumble{
			Context: k8s.Context(server.Name),
			Tag:     "1.1.1",
		},
		"ip": &app.Ip{
			Context: k8s.Context(server.Name),
			IP:      server.IP,
			Tag:     "1.1.0",
			Domain:  "ip.benjamin-borbe.de",
		},
		"password": &app.Password{
			Context: k8s.Context(server.Name),
			Tag:     "1.1.0",
			Domains: k8s.IngressHosts{
				"password.benjamin-borbe.de",
			},
		},
		"now": &app.Now{
			Context: k8s.Context(server.Name),
			Tag:     "1.3.0",
			Domains: k8s.IngressHosts{
				"now.benjamin-borbe.de",
			},
		},
		"helloworld": &app.HelloWorld{
			Context: k8s.Context(server.Name),
			Tag:     "1.0.1",
			Domains: k8s.IngressHosts{
				"rocketsource.de",
				"www.rocketsource.de",
				"rocketnews.de",
				"www.rocketnews.de",
			},
		},
		"slideshow": &app.Slideshow{
			Context: k8s.Context(server.Name),
			Domains: k8s.IngressHosts{
				"slideshow.benjamin-borbe.de",
			},
		},
		"kickstart": &app.Kickstart{
			Context: k8s.Context(server.Name),
			Domains: k8s.IngressHosts{
				"kickstart.benjamin-borbe.de",
				"ks.benjamin-borbe.de",
			},
		},
	}
}

func (w *World) rasp() map[AppName]world.Configuration {
	server := server.Rasp
	return map[AppName]world.Configuration{
		"openvpn-client": &openvpn.RemoteClient{
			SSH: &ssh.SSH{
				Host: ssh.Host{
					IP:   server.IP,
					Port: 22,
				},
				User:           "bborbe",
				PrivateKeyPath: "/Users/bborbe/.ssh/id_rsa",
			},
			ClientName:    openvpn.ClientName(server.Name),
			ServerName:    "hetzner",
			ServerAddress: "hetzner-1.benjamin-borbe.de",
		},
	}
}

func (w *World) co2hz() map[AppName]world.Configuration {
	server := server.Co2hz
	return map[AppName]world.Configuration{
		"openvpn-client": &openvpn.RemoteClient{
			SSH: &ssh.SSH{
				Host: ssh.Host{
					IP:   server.IP,
					Port: 22,
				},
				User:           "bborbe",
				PrivateKeyPath: "/Users/bborbe/.ssh/id_rsa",
			},
			ClientName:    openvpn.ClientName(server.Name),
			ServerName:    "hetzner",
			ServerAddress: "hetzner-1.benjamin-borbe.de",
		},
	}
}

func (w *World) co2wz() map[AppName]world.Configuration {
	server := server.Co2wz
	return map[AppName]world.Configuration{
		"openvpn-client": &openvpn.RemoteClient{
			SSH: &ssh.SSH{
				Host: ssh.Host{
					IP:   server.IP,
					Port: 22,
				},
				User:           "bborbe",
				PrivateKeyPath: "/Users/bborbe/.ssh/id_rsa",
			},
			ClientName:    openvpn.ClientName(server.Name),
			ServerName:    "hetzner",
			ServerAddress: "hetzner-1.benjamin-borbe.de",
		},
	}
}
