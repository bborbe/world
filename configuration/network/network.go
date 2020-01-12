// Copyright (c) 2019 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package network

import (
	"github.com/bborbe/world/configuration/openvpn"
	"github.com/bborbe/world/pkg/network"
)

type Network struct {
	Name  string
	IPNet network.IPNet
}

var PN = Network{
	Name:  "pn",
	IPNet: network.IPNetStatic("192.168.2.0/24"),
}

var HetznerVPN = Network{
	Name:  "hetzner-vpn",
	IPNet: network.IPNetStatic("172.16.90.1/24"),
}

var RaspVPN = Network{
	Name:  "rasp-vpn",
	IPNet: network.IPNetStatic("172.16.80.1/24"),
}

var HM = Network{
	Name:  "hm",
	IPNet: network.IPNetStatic("192.168.178.0/24"),
}

var VPNServer = struct {
	ServerName    openvpn.ServerName
	ServerAddress openvpn.ServerAddress
	Port          network.Port
}{
	ServerName:    "hetzner",
	ServerAddress: "hetzner-1.benjamin-borbe.de",
	Port:          network.PortStatic(563),
}
