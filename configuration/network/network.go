// Copyright (c) 2019 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package network

import (
	"github.com/bborbe/world/pkg/network"
)

const (
	NetcupIP = network.IPStatic("185.170.112.48")

	PNNetwork = network.IPNetStatic("192.168.2.0/24")
	SunIP     = network.IPStatic("192.168.2.3")

	HetznerVPNIPNet = network.IPNetStatic("172.16.90.1/24")

	RaspVPNIPNet = network.IPNetStatic("172.16.80.1/24")

	HMNetwork = network.IPNetStatic("192.168.178.0/24")
	FireIP    = network.IPStatic("192.168.178.3")
	NukeIP    = network.IPStatic("192.168.178.5")
	NovaIP    = network.IPStatic("192.168.178.122")
)
