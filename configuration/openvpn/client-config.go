// Copyright (c) 2019 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package openvpn

import (
	"context"
	"os"
	"os/user"
	"path"
	"path/filepath"

	"github.com/bborbe/world/pkg/content"
	"github.com/bborbe/world/pkg/file"
	"github.com/bborbe/world/pkg/template"
	"github.com/bborbe/world/pkg/validation"
	"github.com/pkg/errors"
)

type ClientConfig struct {
	ClientName   ClientName
	ServerConfig ServerConfig
}

func (c ClientConfig) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		c.ServerConfig.ServerName,
		c.ClientName,
	)
}

func (c *ClientConfig) ConfigContent() content.HasContent {
	return content.Func(func(ctx context.Context) ([]byte, error) {
		type Route struct {
			Gateway string
			IP      string
			Netmask string
		}
		data := struct {
			ServerHost string
			ServerPort int
			Routes     []Route
		}{
			ServerHost: c.ServerConfig.ServerName.String(),
			ServerPort: 563,
			Routes:     []Route{},
		}
		return template.Render(`
#viscosity startonopen true
#viscosity usepeerdns false
#viscosity ipv6 false
#viscosity dns off
#viscosity protocol openvpn
#viscosity autoreconnect true
#viscosity dnssupport true
#viscosity name Home
#viscosity dhcp false
remote {{.ServerHost}} {{.ServerPort}} tcp-client
dev tap
persist-tun
persist-key
compress lzo
pull
tls-client
ca ca.crt
cert cert.crt
key key.key
{{range $route := .Routes}}
route {{$route.Net}} {{$route.Mask}} default default
{{ end }} 
tls-auth ta.key 1
mute-replay-warnings
ns-cert-type server
resolv-retry infinite
comp-lzo adaptive
`, data)
	})
}

func (c *ClientConfig) LocalPath(filename string) file.HasPath {
	return file.PathFunc(func(ctx context.Context) (string, error) {
		directory, err := c.clientDirectory()
		if err != nil {
			return "", err
		}
		return path.Join(directory, filename), nil
	})
}

func (c *ClientConfig) clientDirectory() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", errors.Wrap(err, "get homedir failed")
	}
	dir := filepath.Join(usr.HomeDir, ".openvpn", c.ClientName.String())
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return dir, nil
}
