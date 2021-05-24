// Copyright (c) 2020 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service

import (
	"context"
	"github.com/bborbe/world/pkg/content"
	"github.com/bborbe/world/pkg/deployer"
	"github.com/bborbe/world/pkg/template"
	"github.com/pkg/errors"

	"github.com/bborbe/world/pkg/apt"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type FritzBoxRestart struct {
	SSH              *ssh.SSH
	FritzBoxUser     deployer.SecretValue
	FritzBoxPassword deployer.SecretValue
}

func (d *FritzBoxRestart) Children() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguraionBuilder().WithApplier(&apt.Install{
			SSH:     d.SSH,
			Package: "curl",
		}),
		&Cron{
			SSH:        d.SSH,
			Name:       "fritzbox-restart",
			Expression: d.cronExpression(),
			Schedule:   "0 3 * * *",
		},
	}
}

func (d *FritzBoxRestart) cronExpression() content.HasContent {
	return content.Func(func(ctx context.Context) ([]byte, error) {
		user, err := d.FritzBoxUser.Value(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "get user failed")
		}
		password, err := d.FritzBoxPassword.Value(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "get password failed")
		}

		return template.Render(`curl -s -k -m 5 --anyauth -u "{{.User}}:{{.Password}}" "http://{{.IP}}:49000{{.Location}}" -H 'Content-Type: text/xml; charset="utf-8"' -H "SoapAction:{{.Uri}}#{{.Action}}" -d "<?xml version='1.0' encoding='utf-8'?><s:Envelope s:encodingStyle='http://schemas.xmlsoap.org/soap/encoding/' xmlns:s='http://schemas.xmlsoap.org/soap/envelope/'><s:Body><u:{{.Action}} xmlns:u='{{.Uri}}'></u:{{.Action}}></s:Body></s:Envelope>" > /dev/null`, struct {
			User     string
			Password string
			IP       string
			Location string
			Uri      string
			Action   string
		}{
			User:     string(user),
			Password: string(password),
			IP:       "192.168.178.1",
			Location: "/upnp/control/deviceconfig",
			Uri:      "urn:dslforum-org:service:DeviceConfig:1",
			Action:   "Reboot",
		})
	})
}

func (d *FritzBoxRestart) Applier() (world.Applier, error) {
	return nil, nil
}

func (d *FritzBoxRestart) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		d.SSH,
		d.FritzBoxUser,
		d.FritzBoxPassword,
	)
}
