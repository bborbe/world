// Copyright (c) 2019 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import "github.com/bborbe/world/pkg/network"

type Server struct {
	Name string
	IP   network.IP
}

var Hetzner = Server{
	Name: "hetzner",
	IP:   network.IPStatic("159.69.203.89"),
}

var Netcup = Server{
	Name: "netcup",
	IP:   network.IPStatic("185.170.112.48"),
}

var Sun = Server{
	Name: "sun",
	IP:   network.IPStatic("192.168.2.3"),
}

var Rasp = Server{
	Name: "rasp",
	IP:   network.IPStatic("192.168.178.2"),
}

var Fire = Server{
	Name: "fire",
	IP:   network.IPStatic("192.168.178.3"),
}

var Nuke = Server{
	Name: "nuke",
	IP:   network.IPStatic("192.168.178.5"),
}

var Co2hz = Server{
	Name: "co2hz",
	IP:   network.IPStatic("192.168.178.6"),
}

var Co2wz = Server{
	Name: "co2wz",
	IP:   network.IPStatic("192.168.178.7"),
}

var Star = Server{
	Name: "star",
	IP:   network.IPStatic("192.168.178.101"),
}

var Nova = Server{
	Name: "nova",
	IP:   network.IPStatic("192.168.178.122"),
}
