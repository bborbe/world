// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package configuration

import (
	"context"
	"fmt"

	"github.com/bborbe/world/configuration/service"
	"github.com/bborbe/world/pkg/dns"
	"github.com/bborbe/world/pkg/hetzner"
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

func (w *World) Children(ctx context.Context) (world.Configurations, error) {
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
	return result, nil
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
		"hetzner-1": w.hetzner1(),
		"rasp4":     w.rasp4(),
		// vpn client nodes
		"nuke":            w.vpnClientNode(Nuke),
		"fire":            w.vpnClientNode(Fire),
		"fire-k3s-master": w.vpnClientNode(FireK3sMaster),
		"fire-k3s-prod":   w.vpnClientNode(FireK3sProd),
		"fire-k3s-dev":    w.vpnClientNode(FireK3sDev),
		"hell":            w.vpnClientNode(Hell),
		"rasp3":           w.vpnClientNode(Rasp3),
		"co2hz":           w.vpnClientNode(Co2hz),
		"co2wz":           w.vpnClientNode(Co2wz),
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
		PrivateKeyPath: "/Users/bborbe/.ssh/id_ed25519_personal",
	}
	openvpnClients := []Server{
		Rasp3,
		Rasp4,
		Fire,
		Nuke,
		FireK3sMaster,
		FireK3sProd,
		FireK3sDev,
		Co2hz,
		Co2wz,
		Nova,
		Star,
		Hell,
	}
	return map[AppName]world.Configuration{
		"screego": &service.Screego{
			SSH:     ssh,
			IP:      ip,
			Version: "1.10.3", // https://hub.docker.com/r/screego/server/tags
		},
		"bind": &service.Bind{
			SSH: ssh,
			IP:  ip,
		},
		"ntpdate": &service.NtpDate{
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
			IP:           ip,
			Domain:       IPHostname,
			Target:       "http://localhost:8000",
			Requirements: buildDNSRequirements(ip, IPHostname),
		},
		"teamvault-proxy": &service.NginxProxy{
			SSH:          ssh,
			IP:           ip,
			Domain:       TeamvaultHostname,
			Target:       fmt.Sprintf("http://%s:8000", Sun.VpnIP),
			Requirements: buildDNSRequirements(ip, TeamvaultHostname),
		},
		"screego-proxy": &service.NginxProxy{
			SSH:              ssh,
			IP:               ip,
			Domain:           ScreegoHostname,
			Target:           "http://127.0.0.1:5050",
			Requirements:     buildDNSRequirements(ip, ScreegoHostname),
			WebsocketEnabled: true,
		},
		"ip": &service.Ip{
			SSH:  ssh,
			Tag:  "1.1.0",
			Port: network.PortStatic(8000),
		},
		"poste-proxy": &service.NginxProxy{
			SSH:          ssh,
			IP:           ip,
			Domain:       MailHostname,
			Target:       "http://localhost:8001",
			Requirements: buildDNSRequirements(ip, MailHostname),
		},
		"poste": &service.Poste{
			SSH:          ssh,
			PosteVersion: "2.3.21", // https://hub.docker.com/r/analogic/poste.io/tags
			Port:         network.PortStatic(8001),
		},
	}
}

func (w *World) vpnClientNode(server Server) map[AppName]world.Configuration {
	ssh := &ssh.SSH{
		Host: ssh.Host{
			IP:   server.IP,
			Port: 22,
		},
		User:           "bborbe",
		PrivateKeyPath: "/Users/bborbe/.ssh/id_ed25519_personal",
	}
	return map[AppName]world.Configuration{
		"ntpdate": &service.NtpDate{
			SSH: ssh,
		},
		"openvpn-client": &openvpn.RemoteClient{
			SSH:           ssh,
			ClientName:    openvpn.ClientName(server.Name),
			ServerName:    HetznerVPNServer.ServerName,
			ServerAddress: HetznerVPNServer.ServerAddress,
			ServerPort:    HetznerVPNServer.Port,
			Routes:        BuildRoutes(),
			Device:        openvpn.Tun,
		},
	}
}

func (w *World) rasp4() map[AppName]world.Configuration {
	rasp4 := Rasp4
	ip := rasp4.IP
	ssh := &ssh.SSH{
		Host: ssh.Host{
			IP:   ip,
			Port: 22,
		},
		User:           "bborbe",
		PrivateKeyPath: "/Users/bborbe/.ssh/id_ed25519_personal",
	}
	return map[AppName]world.Configuration{
		"fritzbox-restart": &service.FritzBoxRestart{
			SSH:              ssh,
			FritzBoxUser:     w.TeamvaultSecrets.Username("7qGGQq"),
			FritzBoxPassword: w.TeamvaultSecrets.Password("7qGGQq"),
		},
		"dns-update-home.benjamin-borbe.de": &service.DnsUpdate{
			SSH:        ssh,
			DnsKey:     w.TeamvaultSecrets.File("9L64w3"),
			DnsPrivate: w.TeamvaultSecrets.File("aL50O8"),
			DnsName:    "home",
			DnsZone:    "benjamin-borbe.de",
		},
		"dns-update-home.rocketnews.de": &service.DnsUpdate{
			SSH:        ssh,
			DnsKey:     w.TeamvaultSecrets.File("9L64w3"),
			DnsPrivate: w.TeamvaultSecrets.File("aL50O8"),
			DnsName:    "home",
			DnsZone:    "rocketnews.de",
		},
		"ntpdate": &service.NtpDate{
			SSH: ssh,
		},
		"openvpn-client": &openvpn.RemoteClient{
			SSH:           ssh,
			ClientName:    openvpn.ClientName(rasp4.Name),
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
		result = append(result, buildDNSRequirement(ip, host))
	}
	return result
}

func buildDNSRequirement(ip network.IP, host network.Host) world.Configuration {
	return world.NewConfiguraionBuilder().WithApplier(
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
	)
}
