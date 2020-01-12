// Copyright (c) 2019 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import "github.com/bborbe/world/pkg/network"

type Server struct {
	Name  string
	IP    network.IP
	VpnIP network.IP
}

var Hetzner = Server{
	Name:  "hetzner",
	IP:    network.IPStatic("159.69.203.89"),
	VpnIP: network.IPStatic("172.16.90.1"),
}

var Netcup = Server{
	Name:  "netcup",
	IP:    network.IPStatic("185.170.112.48"),
	VpnIP: network.IPStatic("172.16.90.10"),
}

var Sun = Server{
	Name:  "sun",
	IP:    network.IPStatic("192.168.2.3"),
	VpnIP: network.IPStatic("172.16.90.5"),
}

var Rasp = Server{
	Name:  "rasp",
	IP:    network.IPStatic("192.168.178.2"),
	VpnIP: network.IPStatic("172.16.90.8"),
}

var Fire = Server{
	Name:  "fire",
	IP:    network.IPStatic("192.168.178.3"),
	VpnIP: network.IPStatic("172.16.90.3"),
}

var Nuke = Server{
	Name:  "nuke",
	IP:    network.IPStatic("192.168.178.5"),
	VpnIP: network.IPStatic("172.16.90.4"),
}

var Co2hz = Server{
	Name:  "co2hz",
	IP:    network.IPStatic("192.168.178.6"),
	VpnIP: network.IPStatic("172.16.90.6"),
}

var Co2wz = Server{
	Name:  "co2wz",
	IP:    network.IPStatic("192.168.178.7"),
	VpnIP: network.IPStatic("172.16.90.7"),
}

var Star = Server{
	Name:  "star",
	IP:    network.IPStatic("192.168.178.101"),
	VpnIP: network.IPStatic("172.16.90.9"),
}

var Nova = Server{
	Name:  "nova",
	IP:    network.IPStatic("192.168.178.122"),
	VpnIP: network.IPStatic("172.16.90.2"),
}
