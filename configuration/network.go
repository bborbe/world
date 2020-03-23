// Copyright (c) 2019 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package configuration

import (
	"github.com/bborbe/world/pkg/network"
	"github.com/bborbe/world/pkg/openvpn"
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

var NetcupVPN = Network{
	Name:  "netcup-vpn",
	IPNet: network.IPNetStatic("172.16.80.1/24"),
}

var HM = Network{
	Name:  "hm",
	IPNet: network.IPNetStatic("192.168.178.0/24"),
}

type VPNServer struct {
	ServerName    openvpn.ServerName
	ServerAddress openvpn.ServerAddress
	Port          network.Port
}

var HetznerVPNServer = VPNServer{
	ServerName:    "hetzner",
	ServerAddress: "hetzner-1.benjamin-borbe.de",
	Port:          network.PortStatic(563),
}

var NetcupVPNServer = VPNServer{
	ServerName:    "netcup",
	ServerAddress: "v22016124049440903.goodsrv.de",
	Port:          network.PortStatic(563),
}
