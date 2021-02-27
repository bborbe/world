// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package configuration

import (
	"context"
	"fmt"

	"github.com/bborbe/world/configuration/app"
	"github.com/bborbe/world/configuration/backup"
	"github.com/bborbe/world/configuration/cert_manager"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/ingress_nginx"
	"github.com/bborbe/world/configuration/service"
	"github.com/bborbe/world/pkg/dns"
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
	apiKey := w.TeamvaultSecrets.Password("kLolmq")
	ip := &hetzner.IP{
		Client: w.HetznerClient,
		ApiKey: apiKey,
		Name:   "hetzner-1",
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
		"ubuntu-unattended-upgrades": &service.UbuntuUnattendedUpgrades{
			SSH: ssh,
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
		"nginx": &service.Nginx{
			SSH: ssh,
		},
		"ip-proxy": &service.NginxProxy{
			SSH:          ssh,
			Domain:       IPHostname,
			Target:       "http://localhost:8080",
			Requirements: buildDNSRequirements(ip, IPHostname),
		},
		"teamvault-proxy": &service.NginxProxy{
			SSH:          ssh,
			Domain:       TeamvaultHostname,
			Target:       fmt.Sprintf("http://%s:8000", Sun.VpnIP),
			Requirements: buildDNSRequirements(ip, TeamvaultHostname),
		},
		"confluence-proxy": &service.NginxProxy{
			SSH:          ssh,
			Domain:       ConfluenceHostname,
			Target:       fmt.Sprintf("http://%s:8002", Sun.VpnIP),
			Requirements: buildDNSRequirements(ip, ConfluenceHostname),
		},
		"webdav-proxy": &service.NginxProxy{
			SSH:          ssh,
			Domain:       WebdavHostname,
			Target:       fmt.Sprintf("http://%s:8004", Sun.VpnIP),
			Requirements: buildDNSRequirements(ip, WebdavHostname),
		},
		"ip": &service.Ip{
			SSH: ssh,
			Tag: "1.1.0",
			Port: network.PortStatic(8000),
		},
		"poste-proxy": &service.NginxProxy{
			SSH:          ssh,
			Domain:       MailHostname,
			Target:       "http://localhost:8001",
			Requirements: buildDNSRequirements(ip, MailHostname),
		},
		"poste": &service.Poste{
			SSH:          ssh,
			PosteVersion: "2.2.29", // https://hub.docker.com/r/analogic/poste.io/tags
			Port:         network.PortStatic(8001),
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
	ip := sun.IP
	ssh := &ssh.SSH{
		Host: ssh.Host{
			IP:   ip,
			Port: 22,
		},
		User:           "bborbe",
		PrivateKeyPath: "/Users/bborbe/.ssh/id_rsa",
	}
	return map[AppName]world.Configuration{
		"cluster": &cluster.Sun{
			SSH:       ssh,
			Context:   k8s.Context(sun.Name),
			ClusterIP: ip,
		},
		"cluster-admin": &service.ClusterAdmin{
			Context: k8s.Context(sun.Name),
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
		"ldap": &service.Ldap{
			SSH:          ssh,
			Tag:          "1.3.0",
			LdapPassword: w.TeamvaultSecrets.Password("MOPMLG"),
		},
		"teamvault": &service.Teamvault{
			SSH:              ssh,
			AppPort:          network.PortStatic(8000),
			DBPort:           network.PortStatic(8001),
			Domain:           "teamvault.benjamin-borbe.de",
			DatabasePassword: w.TeamvaultSecrets.Password("VO0W5w"),
			SmtpUsername:     w.TeamvaultSecrets.Username("3OlNaq"),
			SmtpPassword:     w.TeamvaultSecrets.Password("3OlNaq"),
			LdapPassword:     w.TeamvaultSecrets.Password("MOPMLG"),
			SecretKey:        w.TeamvaultSecrets.Password("NqA68w"),
			FernetKey:        w.TeamvaultSecrets.Password("5wYZ2O"),
			Salt:             w.TeamvaultSecrets.Password("Rwg74w"),
		},
		"confluence": &service.Confluence{
			SSH:              ssh,
			AppPort:          network.PortStatic(8002),
			DBPort:           network.PortStatic(8003),
			Domain:           ConfluenceHostname,
			Version:          "7.11.1",
			DatabasePassword: w.TeamvaultSecrets.Password("3OlaLn"),
			SmtpUsername:     w.TeamvaultSecrets.Username("nOeNjL"),
			SmtpPassword:     w.TeamvaultSecrets.Password("nOeNjL"),
		},
		"webdav": &service.Webdav{
			SSH:            ssh,
			Port:           network.PortStatic(8004),
			WebdavPassword: w.TeamvaultSecrets.Password("VOzvAO"),
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
		"backup": &app.BackupServer{
			Context: k8sContext,
		},
		"poste": &app.Poste{
			Context:      k8sContext,
			PosteVersion: "2.2.29", // https://hub.docker.com/r/analogic/poste.io/tags
			Domains: k8s.IngressHosts{
				k8s.IngressHost(MailHostname),
			},
			Requirements: buildDNSRequirements(ip, MailHostname),
		},
		"bind": &app.Bind{
			Context: k8sContext,
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
