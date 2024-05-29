// Copyright (c) 2019 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package configuration

import "github.com/bborbe/world/pkg/network"

type Server struct {
	Name    string
	IP      network.IP
	IPMask  network.IPMask
	VpnIP   network.IP
	Gateway network.IP
}

var Hetzner = Server{
	Name:    "hetzner",
	IP:      network.IPStatic("159.69.203.89"),
	IPMask:  network.MaskStatic(32),
	Gateway: network.IPStatic("172.31.1.1"),
	VpnIP:   network.IPStatic("172.16.90.1"),
}

var Nova = Server{
	Name:    "nova",
	IP:      network.IPStatic("192.168.178.122"),
	IPMask:  network.MaskStatic(24),
	Gateway: network.IPStatic("192.168.178.1"),
	VpnIP:   network.IPStatic("172.16.90.16"),
}

var Fire = Server{
	Name:    "fire",
	IP:      network.IPStatic("192.168.178.3"),
	IPMask:  network.MaskStatic(24),
	Gateway: network.IPStatic("192.168.178.1"),
	VpnIP:   network.IPStatic("172.16.90.8"),
}

var Hell = Server{
	Name:    "hell",
	IP:      network.IPStatic("192.168.178.9"),
	IPMask:  network.MaskStatic(24),
	Gateway: network.IPStatic("192.168.178.1"),
	VpnIP:   network.IPStatic("172.16.90.33"),
}

var Nuke = Server{
	Name:    "nuke",
	IP:      network.IPStatic("192.168.178.5"),
	IPMask:  network.MaskStatic(24),
	Gateway: network.IPStatic("192.168.178.1"),
	VpnIP:   network.IPStatic("172.16.90.4"),
}

var Sun = Server{
	Name:    "sun",
	IP:      network.IPStatic("192.168.2.3"),
	IPMask:  network.MaskStatic(24),
	Gateway: network.IPStatic("192.168.2.1"),
	VpnIP:   network.IPStatic("172.16.90.12"),
}

var Co2hz = Server{
	Name:    "co2hz",
	IP:      network.IPStatic("192.168.178.6"),
	IPMask:  network.MaskStatic(24),
	Gateway: network.IPStatic("192.168.178.1"),
	VpnIP:   network.IPStatic("172.16.90.28"),
}

var Co2wz = Server{
	Name:    "co2wz",
	IP:      network.IPStatic("192.168.178.7"),
	IPMask:  network.MaskStatic(24),
	Gateway: network.IPStatic("192.168.178.1"),
	VpnIP:   network.IPStatic("172.16.90.24"),
}

var Rasp3 = Server{
	Name:    "rasp3",
	IP:      network.IPStatic("192.168.178.2"),
	IPMask:  network.MaskStatic(24),
	Gateway: network.IPStatic("192.168.178.1"),
	VpnIP:   network.IPStatic("172.16.90.20"),
}

var Rasp4 = Server{
	Name:    "rasp4",
	IP:      network.IPStatic("192.168.178.8"),
	IPMask:  network.MaskStatic(24),
	Gateway: network.IPStatic("192.168.178.1"),
	VpnIP:   network.IPStatic("172.16.90.29"),
}

var Star = Server{
	Name:    "star",
	IP:      network.IPStatic("192.168.178.101"),
	IPMask:  network.MaskStatic(24),
	Gateway: network.IPStatic("192.168.178.1"),
	VpnIP:   network.IPStatic("172.16.90.32"),
}

var FireK3sMaster = Server{
	Name:    "fire-k3s-master",
	IP:      network.IPStatic("192.168.178.20"),
	IPMask:  network.MaskStatic(24),
	Gateway: network.IPStatic("192.168.178.1"),
	VpnIP:   network.IPStatic("172.16.90.3"),
}

var FireK3sProd = Server{
	Name:    "fire-k3s-prod",
	IP:      network.IPStatic("192.168.178.21"),
	IPMask:  network.MaskStatic(24),
	Gateway: network.IPStatic("192.168.178.1"),
	VpnIP:   network.IPStatic("172.16.90.4"),
}

var FireK3sDev = Server{
	Name:    "fire-k3s-dev",
	IP:      network.IPStatic("192.168.178.22"),
	IPMask:  network.MaskStatic(24),
	Gateway: network.IPStatic("192.168.178.1"),
	VpnIP:   network.IPStatic("172.16.90.5"),
}
