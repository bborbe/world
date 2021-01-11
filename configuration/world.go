// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package configuration

import (
	"context"

	"github.com/bborbe/world/configuration/app"
	"github.com/bborbe/world/configuration/backup"
	"github.com/bborbe/world/configuration/cert_manager"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/ingress_nginx"
	"github.com/bborbe/world/configuration/service"
	"github.com/bborbe/world/pkg/dns"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/hetzner"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/network"
	"github.com/bborbe/world/pkg/openvpn"
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
		"rasp3":     w.rasp3(),
		"rasp4":     w.rasp4(),
		"co2hz":     w.co2hz(),
		"co2wz":     w.co2wz(),
		"star":      w.star(),
	}
}

func (w *World) hetzner1() map[AppName]world.Configuration {
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

	openvpnClients := []Server{
		Sun,
		Rasp3,
		Rasp4,
		Fire,
		Nuke,
		Co2hz,
		Co2wz,
		Nova,
		Star,
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
		"openvpn-net": &openvpn.Server{
			SSH:         ssh,
			ServerName:  HetznerVPNServer.ServerName,
			ServerIPNet: HetznerVPN.IPNet,
			ServerPort:  HetznerVPNServer.Port,
			IRoutes:     BuildIRoutes(openvpnClients...),
			ClientIPs:   BuildClientIPs(openvpnClients...),
			Device:      openvpn.Tun,
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
		"ingress-nginx": &ingress_nginx.App{
			Context: k8sContext,
		},
		"cert-manager": &cert_manager.App{
			Context: k8sContext,
		},
	}
}

func (w *World) nova() map[AppName]world.Configuration {
	nova := Nova
	return map[AppName]world.Configuration{
		"cluster": &cluster.Nova{
			IP:   nova.IP,
			Host: "nova.hm.benjamin-borbe.de",
		},
		"openvpn-client": &openvpn.LocalClient{
			ClientName:    openvpn.ClientName(nova.Name),
			ServerName:    HetznerVPNServer.ServerName,
			ServerAddress: HetznerVPNServer.ServerAddress,
			ServerPort:    HetznerVPNServer.Port,
			Routes:        BuildRoutes(),
			Device:        openvpn.Tun,
		},
	}
}

func (w *World) star() map[AppName]world.Configuration {
	star := Star
	return map[AppName]world.Configuration{
		"openvpn-client": &openvpn.LocalClient{
			ClientName:    openvpn.ClientName(star.Name),
			ServerName:    HetznerVPNServer.ServerName,
			ServerAddress: HetznerVPNServer.ServerAddress,
			ServerPort:    HetznerVPNServer.Port,
			Routes:        BuildRoutes(),
			Device:        openvpn.Tun,
		},
	}
}

func (w *World) fire() map[AppName]world.Configuration {
	fire := Fire
	ssh := &ssh.SSH{
		Host: ssh.Host{
			IP:   fire.IP,
			Port: 22,
		},
		User:           "bborbe",
		PrivateKeyPath: "/Users/bborbe/.ssh/id_rsa",
	}
	return map[AppName]world.Configuration{
		"cluster": &cluster.Fire{
			SSH:       ssh,
			Context:   k8s.Context(fire.Name),
			ClusterIP: fire.IP,
		},
		"cluster-admin": &service.ClusterAdmin{
			Context: k8s.Context(fire.Name),
		},
		"calico": &service.Calico{
			Context:   k8s.Context(fire.Name),
			ClusterIP: fire.IP,
		},
		"dns": &app.CoreDns{
			Context: k8s.Context(fire.Name),
		},
		"openvpn-client": &openvpn.RemoteClient{
			SSH:           ssh,
			ClientName:    openvpn.ClientName(fire.Name),
			ServerName:    HetznerVPNServer.ServerName,
			ServerAddress: HetznerVPNServer.ServerAddress,
			ServerPort:    HetznerVPNServer.Port,
			Routes: BuildRoutes(
				Sun,
			),
			Device: openvpn.Tun,
		},
		"ingress-nginx": &ingress_nginx.App{
			Context: k8s.Context(fire.Name),
		},
		"backup": &app.BackupClient{
			Context: k8s.Context(fire.Name),
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
	nuke := Nuke
	ssh := &ssh.SSH{
		Host: ssh.Host{
			IP:   nuke.IP,
			Port: 22,
		},
		User:           "bborbe",
		PrivateKeyPath: "/Users/bborbe/.ssh/id_rsa",
	}
	return map[AppName]world.Configuration{
		"cluster": &cluster.Nuke{
			SSH:       ssh,
			Context:   k8s.Context(nuke.Name),
			ClusterIP: nuke.IP,
		},
		"cluster-admin": &service.ClusterAdmin{
			Context: k8s.Context(nuke.Name),
		},
		"calico": &service.Calico{
			Context:   k8s.Context(nuke.Name),
			ClusterIP: nuke.IP,
		},
		"dns": &app.CoreDns{
			Context: k8s.Context(nuke.Name),
		},
		"ingress-nginx": &ingress_nginx.App{
			Context: k8s.Context(nuke.Name),
		},
		"openvpn-client": &openvpn.RemoteClient{
			SSH:           ssh,
			ClientName:    openvpn.ClientName(nuke.Name),
			ServerName:    HetznerVPNServer.ServerName,
			ServerAddress: HetznerVPNServer.ServerAddress,
			ServerPort:    HetznerVPNServer.Port,
			Routes:        BuildRoutes(),
			Device:        openvpn.Tun,
		},
		"backup": &app.BackupClient{
			Context: k8s.Context(nuke.Name),
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
	sun := Sun
	ssh := &ssh.SSH{
		Host: ssh.Host{
			IP:   sun.IP,
			Port: 22,
		},
		User:           "bborbe",
		PrivateKeyPath: "/Users/bborbe/.ssh/id_rsa",
	}
	return map[AppName]world.Configuration{
		"cluster": &cluster.Sun{
			SSH:       ssh,
			Context:   k8s.Context(sun.Name),
			ClusterIP: sun.IP,
		},
		"cluster-admin": &service.ClusterAdmin{
			Context: k8s.Context(sun.Name),
		},
		"calico": &service.Calico{
			Context:   k8s.Context(sun.Name),
			ClusterIP: sun.IP,
		},
		"dns": &app.CoreDns{
			Context: k8s.Context(sun.Name),
		},
		"hostpath": &app.HostPathProvisioner{
			Context:             k8s.Context(sun.Name),
			HostPath:            "/data/hostpath-provisioner",
			DefaultStorageClass: true,
		},
		"openvpn-client": &openvpn.RemoteClient{
			SSH:           ssh,
			ClientName:    openvpn.ClientName(sun.Name),
			ServerName:    HetznerVPNServer.ServerName,
			ServerAddress: HetznerVPNServer.ServerAddress,
			ServerPort:    HetznerVPNServer.Port,
			Routes: BuildRoutes(
				Co2hz,
				Co2wz,
				Rasp3,
				Rasp4,
				Fire,
				Nuke,
			),
			Device: openvpn.Tun,
		},
		"minecraft": &app.Minecraft{
			Context: k8s.Context(sun.Name),
		},
		"monitoring": &app.Monitoring{
			Context:         k8s.Context(sun.Name),
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
		"ingress-nginx": &ingress_nginx.App{
			Context: k8s.Context(sun.Name),
		},
		"backup": &app.BackupClient{
			Context: k8s.Context(sun.Name),
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
		"mqtt-kafka-connector-co2mon": &app.MqttKafkaConnector{
			Context:      k8s.Context(sun.Name),
			MqttBroker:   "tcp://rasp3.hm.benjamin-borbe.de:1883",
			MqttUser:     w.TeamvaultSecrets.Username("9qNx3O"),
			MqttPassword: w.TeamvaultSecrets.Password("9qNx3O"),
			MqttTopic:    "co2mon",
			KafkaBrokers: []string{"kafka-cp-kafka-headless.kafka.svc.cluster.local:9092"},
			KafkaTopic:   "co2mon",
		},
		"metabase": &app.Metabase{
			Context:          k8s.Context(sun.Name),
			Domain:           "metabase.benjamin-borbe.de",
			DatabasePassword: w.TeamvaultSecrets.Password("dwkWAw"),
			Requirements:     buildDNSRequirements(sun.IP, MetabaseHostname),
		},
		"kafka": &app.Kafka{
			AccessMode:        "ReadWriteMany",
			Context:           k8s.Context(sun.Name),
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
			Context:      k8s.Context(sun.Name),
			Replicas:     1,
			Domain:       "kafka-status.benjamin-borbe.de",
			Requirements: buildDNSRequirements(sun.IP, KafkaStatus),
		},
		"kafka-latest-versions": &app.KafkaLatestVersions{
			Context:      k8s.Context(sun.Name),
			Replicas:     2,
			AccessMode:   "ReadWriteMany",
			StorageClass: "hostpath",
			Domain:       "versions.benjamin-borbe.de",
			Requirements: buildDNSRequirements(sun.IP, VersionsHostname),
		},
		"kafka-update-available": &app.KafkaUpdateAvailable{
			Context:      k8s.Context(sun.Name),
			Replicas:     2,
			AccessMode:   "ReadWriteMany",
			StorageClass: "hostpath",
			Domain:       "updates.benjamin-borbe.de",
			Requirements: buildDNSRequirements(sun.IP, UpdatesHostname),
		},
		"kafka-k8s-version-collector": &app.KafkaK8sVersionCollector{
			Context: k8s.Context(sun.Name),
		},
		"kafka-dockerhub-version-collector": &app.KafkaDockerhubVersionCollector{
			Context: k8s.Context(sun.Name),
			Repositories: []docker.Repository{
				"confluentinc/cp-kafka-connect",
				"confluentinc/cp-kafka-rest",
				"confluentinc/cp-kafka-rest",
				"confluentinc/cp-kafka",
				"confluentinc/cp-ksql-net",
				"confluentinc/cp-ksql-net",
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
			Context:     k8s.Context(sun.Name),
			Application: "Golang",
			Url:         "https://golang.org/dl/",
			Regex:       `https://dl.google.com/go/go([\d\.]+)\.src\.tar\.gz`,
		},
		"kafka-atlassian-version-collector": &app.KafkaAtlassianVersionCollector{
			Context: k8s.Context(sun.Name),
		},
		"kafka-installed-version-collector": &app.KafkaInstalledVersionCollector{
			Context: k8s.Context(sun.Name),
			Apps: []struct {
				Name  string
				Regex string
				URL   string
			}{
				{
					Name:  "Confluence",
					Regex: `<meta\s+name="ajs-version-number"\s+content="([^"]+)">`,
					URL:   "https://confluence.benjamin-borbe.de",
				},
				{
					Name:  "Jira",
					Regex: `<meta\s+name="ajs-version-number"\s+content="([^"]+)">`,
					URL:   "https://jira.benjamin-borbe.de",
				},
			},
		},
		"kafka-sample": &app.KafkaSample{
			Context:      k8s.Context(sun.Name),
			Domain:       "kafka-sample.benjamin-borbe.de",
			Requirements: buildDNSRequirements(sun.IP, KafkaSampleHostname),
		},
	}
}

func (w *World) netcup() map[AppName]world.Configuration {
	netcup := Netcup
	ip := netcup.IP
	ssh := &ssh.SSH{
		Host: ssh.Host{
			IP:   ip,
			Port: 22,
		},
		User:           ssh.User("bborbe"),
		PrivateKeyPath: "/Users/bborbe/.ssh/id_rsa",
	}
	k8sContext := k8s.Context(netcup.Name)
	return map[AppName]world.Configuration{
		"cluster": &cluster.Netcup{
			SSH:     ssh,
			Context: k8sContext,
			IP:      ip,
		},
		"cluster-admin": &service.ClusterAdmin{
			Context: k8sContext,
		},
		"calico": &service.Calico{
			Context:   k8sContext,
			ClusterIP: ip,
		},
		"hostpath": &app.HostPathProvisioner{
			Context:             k8sContext,
			HostPath:            "/data/hostpath-provisioner",
			DefaultStorageClass: true,
		},
		"dns": &app.CoreDns{
			Context: k8sContext,
		},
		"ingress-nginx": &ingress_nginx.App{
			Context: k8sContext,
		},
		"cert-manager": &cert_manager.App{
			Context: k8sContext,
		},
		"debug": &app.Debug{
			Context: k8sContext,
			Domains: k8s.IngressHosts{
				k8s.IngressHost(DebugHostname.String()),
			},
			Requirements: buildDNSRequirements(ip, DebugHostname),
		},
		"grafana": &app.Grafana{
			Context: k8sContext,
			Domains: k8s.IngressHosts{
				k8s.IngressHost(GrafanaHostname.String()),
			},
			LdapUsername: w.TeamvaultSecrets.Username("MOPMLG"),
			LdapPassword: w.TeamvaultSecrets.Password("MOPMLG"),
			Requirements: buildDNSRequirements(ip, GrafanaHostname),
		},
		"prometheus": &app.Prometheus{
			Context:             k8sContext,
			PrometheusDomains:   k8s.IngressHosts{k8s.IngressHost(PrometheusHostname)},
			AlertmanagerDomains: k8s.IngressHosts{k8s.IngressHost(AlertmanagerHostname)},
			Secret:              w.TeamvaultSecrets.Password("aqMr6w"),
			LdapUsername:        w.TeamvaultSecrets.Username("MOPMLG"),
			LdapPassword:        w.TeamvaultSecrets.Password("MOPMLG"),
			Requirements:        buildDNSRequirements(ip, PrometheusHostname, AlertmanagerHostname),
		},
		"ldap": &app.Ldap{
			Context:      k8sContext,
			Tag:          "1.3.0",
			LdapPassword: w.TeamvaultSecrets.Password("MOPMLG"),
		},
		"teamvault": &app.Teamvault{
			Context: k8sContext,
			Domains: k8s.IngressHosts{
				k8s.IngressHost(TeamvaultHostname.String()),
			},
			DatabasePassword: w.TeamvaultSecrets.Password("VO0W5w"),
			SmtpUsername:     w.TeamvaultSecrets.Username("3OlNaq"),
			SmtpPassword:     w.TeamvaultSecrets.Password("3OlNaq"),
			LdapPassword:     w.TeamvaultSecrets.Password("MOPMLG"),
			SecretKey:        w.TeamvaultSecrets.Password("NqA68w"),
			FernetKey:        w.TeamvaultSecrets.Password("5wYZ2O"),
			Salt:             w.TeamvaultSecrets.Password("Rwg74w"),
			Requirements:     buildDNSRequirements(ip, TeamvaultHostname),
		},
		"monitoring": &app.Monitoring{
			Context:         k8sContext,
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
			},
		},
		"confluence": &app.Confluence{
			Context: k8sContext,
			Domains: k8s.IngressHosts{
				k8s.IngressHost(ConfluenceHostname),
			},
			Version:          "7.9.3",
			DatabasePassword: w.TeamvaultSecrets.Password("3OlaLn"),
			SmtpUsername:     w.TeamvaultSecrets.Username("nOeNjL"),
			SmtpPassword:     w.TeamvaultSecrets.Password("nOeNjL"),
			Requirements:     buildDNSRequirements(ip, ConfluenceHostname),
		},
		"jira": &app.Jira{
			Context: k8sContext,
			Domains: k8s.IngressHosts{
				k8s.IngressHost(JiraHostname),
			},
			Version:          "8.14.0",
			DatabasePassword: w.TeamvaultSecrets.Password("eOB12w"),
			SmtpUsername:     w.TeamvaultSecrets.Username("MwmE0w"),
			SmtpPassword:     w.TeamvaultSecrets.Password("MwmE0w"),
			Requirements:     buildDNSRequirements(ip, JiraHostname),
		},
		"backup": &app.BackupServer{
			Context: k8sContext,
		},
		"poste": &app.Poste{
			Context:      k8sContext,
			PosteVersion: "2.2.26", // https://hub.docker.com/r/analogic/poste.io/tags
			Domains: k8s.IngressHosts{
				k8s.IngressHost(MailHostname),
			},
			Requirements: buildDNSRequirements(ip, MailHostname),
		},
		"maven": &app.Maven{
			Context: k8sContext,
			Domains: k8s.IngressHosts{
				k8s.IngressHost(MavenHostname),
			},
			MavenRepoVersion: "1.0.0",
			Requirements:     buildDNSRequirements(ip, MavenHostname),
		},
		"portfolio": &app.Portfolio{
			Context: k8sContext,
			Domains: k8s.IngressHosts{
				"benjamin-borbe.de",
				"www.benjamin-borbe.de",
				"benjaminborbe.de",
				"www.benjaminborbe.de",
			},
			OverlayServerVersion: "1.0.0",
			GitSyncPassword:      w.TeamvaultSecrets.Password("YLb4wV"),
			Requirements: buildDNSRequirements(
				ip,
				"benjamin-borbe.de",
				"www.benjamin-borbe.de",
				"benjaminborbe.de",
				"www.benjaminborbe.de",
			),
		},
		"webdav": &app.Webdav{
			Context: k8sContext,
			Domains: k8s.IngressHosts{
				"webdav.benjamin-borbe.de",
			},
			Password: w.TeamvaultSecrets.Password("VOzvAO"),
			Requirements: buildDNSRequirements(
				ip,
				"webdav.benjamin-borbe.de",
			),
		},
		"bind": &app.Bind{
			Context: k8sContext,
		},
		"download": &app.Download{
			Context: k8sContext,
			Domains: k8s.IngressHosts{
				"dl.benjamin-borbe.de",
				"download.benjamin-borbe.de",
			},
			Requirements: buildDNSRequirements(
				ip,
				"dl.benjamin-borbe.de",
				"download.benjamin-borbe.de",
			),
		},
		"mumble": &app.Mumble{
			Context: k8sContext,
			Tag:     "1.1.1",
		},
		"password": &app.Password{
			Context: k8sContext,
			Tag:     "1.1.0",
			Domains: k8s.IngressHosts{
				"password.benjamin-borbe.de",
			},
			Requirements: buildDNSRequirements(
				ip,
				"password.benjamin-borbe.de",
			),
		},
		"now": &app.Now{
			Context: k8sContext,
			Tag:     "1.3.0",
			Domains: k8s.IngressHosts{
				"now.benjamin-borbe.de",
			},
			Requirements: buildDNSRequirements(
				ip,
				"now.benjamin-borbe.de",
			),
		},
		"helloworld": &app.HelloWorld{
			Context: k8sContext,
			Tag:     "1.0.1",
			Domains: k8s.IngressHosts{
				"rocketsource.de",
				"www.rocketsource.de",
				"rocketnews.de",
				"www.rocketnews.de",
			},
			Requirements: buildDNSRequirements(
				ip,
				"rocketnews.de",
				"www.rocketnews.de",
			),
		},
		"slideshow": &app.Slideshow{
			Context: k8sContext,
			Domains: k8s.IngressHosts{
				"slideshow.benjamin-borbe.de",
			},
			Requirements: buildDNSRequirements(
				ip,
				"slideshow.benjamin-borbe.de",
			),
		},
		"kickstart": &app.Kickstart{
			Context: k8sContext,
			Domains: k8s.IngressHosts{
				"kickstart.benjamin-borbe.de",
				"ks.benjamin-borbe.de",
			},
			Requirements: buildDNSRequirements(
				ip,
				"kickstart.benjamin-borbe.de",
				"ks.benjamin-borbe.de",
			),
		},
		"ip": &app.Ip{
			Context: k8sContext,
			IP:      ip,
			Tag:     "1.1.0",
			Domains: k8s.IngressHosts{
				k8s.IngressHost(IPHostname),
			},
			Requirements: buildDNSRequirements(ip, IPHostname),
		},
	}
}

func (w *World) rasp3() map[AppName]world.Configuration {
	rasp := Rasp3
	return map[AppName]world.Configuration{
		"openvpn-client": &openvpn.RemoteClient{
			SSH: &ssh.SSH{
				Host: ssh.Host{
					IP:   rasp.IP,
					Port: 22,
				},
				User:           "bborbe",
				PrivateKeyPath: "/Users/bborbe/.ssh/id_rsa",
			},
			ClientName:    openvpn.ClientName(rasp.Name),
			ServerName:    HetznerVPNServer.ServerName,
			ServerAddress: HetznerVPNServer.ServerAddress,
			ServerPort:    HetznerVPNServer.Port,
			Routes:        BuildRoutes(),
			Device:        openvpn.Tun,
		},
	}
}

func (w *World) rasp4() map[AppName]world.Configuration {
	rasp := Rasp4
	return map[AppName]world.Configuration{
		"openvpn-client": &openvpn.RemoteClient{
			SSH: &ssh.SSH{
				Host: ssh.Host{
					IP:   rasp.IP,
					Port: 22,
				},
				User:           "bborbe",
				PrivateKeyPath: "/Users/bborbe/.ssh/id_rsa",
			},
			ClientName:    openvpn.ClientName(rasp.Name),
			ServerName:    HetznerVPNServer.ServerName,
			ServerAddress: HetznerVPNServer.ServerAddress,
			ServerPort:    HetznerVPNServer.Port,
			Routes:        BuildRoutes(),
			Device:        openvpn.Tun,
		},
	}
}

func (w *World) co2hz() map[AppName]world.Configuration {
	co2hz := Co2hz
	return map[AppName]world.Configuration{
		"openvpn-client": &openvpn.RemoteClient{
			SSH: &ssh.SSH{
				Host: ssh.Host{
					IP:   co2hz.IP,
					Port: 22,
				},
				User:           "bborbe",
				PrivateKeyPath: "/Users/bborbe/.ssh/id_rsa",
			},
			ClientName:    openvpn.ClientName(co2hz.Name),
			ServerName:    HetznerVPNServer.ServerName,
			ServerAddress: HetznerVPNServer.ServerAddress,
			ServerPort:    HetznerVPNServer.Port,
			Routes:        BuildRoutes(),
			Device:        openvpn.Tun,
		},
	}
}

func (w *World) co2wz() map[AppName]world.Configuration {
	co2wz := Co2wz
	return map[AppName]world.Configuration{
		"openvpn-client": &openvpn.RemoteClient{
			SSH: &ssh.SSH{
				Host: ssh.Host{
					IP:   co2wz.IP,
					Port: 22,
				},
				User:           "bborbe",
				PrivateKeyPath: "/Users/bborbe/.ssh/id_rsa",
			},
			ClientName:    openvpn.ClientName(co2wz.Name),
			ServerName:    HetznerVPNServer.ServerName,
			ServerAddress: HetznerVPNServer.ServerAddress,
			ServerPort:    HetznerVPNServer.Port,
			Routes:        BuildRoutes(),
			Device:        openvpn.Tun,
		},
	}
}
func buildDNSRequirements(ip network.IP, hosts ...network.Host) []world.Configuration {
	var result []world.Configuration
	for _, host := range hosts {
		result = append(result, world.NewConfiguraionBuilder().WithApplier(
			&dns.Server{
				Host:    "ns.rocketsource.de",
				KeyPath: "/Users/bborbe/.dns/home.benjamin-borbe.de.key",
				List: []dns.Entry{
					{
						Host: host,
						IP:   ip,
					},
				},
			},
		))
	}
	return result
}
