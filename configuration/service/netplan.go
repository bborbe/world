// Copyright (c) 2021 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service

import (
	"context"
	"strconv"

	"github.com/pkg/errors"

	"github.com/bborbe/world/pkg/content"
	"github.com/bborbe/world/pkg/file"
	"github.com/bborbe/world/pkg/network"
	"github.com/bborbe/world/pkg/remote"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/template"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type NetPlan struct {
	SSH     *ssh.SSH
	Gateway network.IP
	IP      network.IP
	IPMask  network.IPMask
	Device  network.Device
}

func (d *NetPlan) Children(ctx context.Context) (world.Configurations, error) {
	return world.Configurations{
		&remote.File{
			SSH:     d.SSH,
			Path:    file.Path("/etc/netplan/00-installer-config.yaml"),
			Content: d.content(),
			User:    "root",
			Group:   "root",
			Perm:    0644,
		},
	}, nil
}

func (d *NetPlan) Applier() (world.Applier, error) {
	return nil, nil
}

func (d *NetPlan) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		d.SSH,
		d.Device,
		d.IP,
		d.IPMask,
		d.Gateway,
	)
}

func (d *NetPlan) content() content.HasContent {
	return content.Func(func(ctx context.Context) ([]byte, error) {

		ip, err := d.IP.IP(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "get ip failed")
		}

		ipMask, err := d.IPMask.IPMask(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "get mask failed")
		}
		ones, _ := ipMask.Size()

		gateway, err := d.Gateway.IP(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "get gateway failed")
		}

		return template.Render(netplanTemplate, struct {
			Gateway string
			IPMask  string
			IP      string
			Device  string
		}{
			Gateway: gateway.String(),
			IPMask:  strconv.Itoa(ones),
			IP:      ip.String(),
			Device:  d.Device.String(),
		})

	})
}

const netplanTemplate = `# This is the network config written by 'world'
network:
  version: 2
  ethernets:
    {{ .Device }}:
      addresses: [{{.IP}}/{{.IPMask}}]
      gateway4: {{.Gateway}}
      nameservers:
        addresses: [8.8.4.4, 8.8.8.8]
`
