// Copyright (c) 2020 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package configuration

import (
	"github.com/bborbe/world/pkg/network"
	"github.com/bborbe/world/pkg/openvpn"
)

func BuildIRoutes(servers ...Server) openvpn.IRoutes {
	var result openvpn.IRoutes
	for _, server := range servers {
		result = append(result, openvpn.IRoute{
			Name: openvpn.ClientName(server.Name),
			IPNet: network.IPNetFromIP{
				IP:   server.IP,
				Mask: 32,
			},
		})
	}
	return result
}

func BuildRoutes(servers ...Server) openvpn.Routes {
	var result openvpn.Routes
	for _, server := range servers {
		result = append(result, openvpn.Route{
			Gateway: server.VpnIP,
			IPNet: network.IPNetFromIP{
				IP:   server.IP,
				Mask: 32,
			},
		})
	}
	return result
}

func BuildClientIPs(servers ...Server) openvpn.ClientIPs {
	var result openvpn.ClientIPs
	for _, server := range servers {
		result = append(result, openvpn.ClientIP{
			Name: openvpn.ClientName(server.Name),
			IP:   server.VpnIP,
		})
	}
	return result
}
