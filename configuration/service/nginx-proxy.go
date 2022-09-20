// Copyright (c) 2020 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service

import (
	"context"
	"fmt"

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

type NginxProxy struct {
	SSH              *ssh.SSH
	Domain           network.Host
	Target           string
	Requirements     []world.Configuration
	WebsocketEnabled bool
	IP               network.IP
}

func (s *NginxProxy) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		s.SSH,
		s.Domain,
		s.IP,
	)
}

func (s *NginxProxy) Children(ctx context.Context) (world.Configurations, error) {
	var result []world.Configuration
	result = append(result, s.Requirements...)
	result = append(result, s.proxy()...)
	return result, nil
}

func (s *NginxProxy) proxy() []world.Configuration {
	return world.Configurations{
		&remote.File{
			SSH:  s.SSH,
			Path: file.Path(fmt.Sprintf("/etc/nginx/sites-enabled/%s.conf", s.Domain)),
			Content: content.Func(func(ctx context.Context) ([]byte, error) {
				ip, err := s.IP.IP(ctx)
				if err != nil {
					return nil, errors.Wrap(err, "get ip failed")
				}

				return template.Render(`
server {
	server_name {{ .Domain }};
	client_max_body_size 100M;

	location / {
{{- if .WebsocketEnabled }} 
		proxy_set_header   Upgrade $http_upgrade;
		proxy_set_header   Connection "upgrade";
		proxy_redirect     http:// $scheme://;
{{- end }}

		proxy_set_header X-Forwarded-Host $host:$server_port;
		proxy_set_header X-Forwarded-Server $host;
		proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
		proxy_set_header X-Forwarded-Proto https;
		proxy_set_header Host $host;
		proxy_pass {{ .Target }}/;
	}

	listen {{ .IP }}:443 ssl;
	ssl_certificate /etc/letsencrypt/live/{{ .Domain }}/fullchain.pem;
	ssl_certificate_key /etc/letsencrypt/live/{{ .Domain }}/privkey.pem;
	include /etc/letsencrypt/options-ssl-nginx.conf;
	ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem;
}

server {
	listen {{ .IP }}:80;

	server_name {{ .Domain }};
    return 301 https://$server_name$request_uri;
}
`, struct {
					Domain           string
					Target           string
					WebsocketEnabled bool
					IP               string
				}{
					Domain:           s.Domain.String(),
					Target:           s.Target,
					WebsocketEnabled: s.WebsocketEnabled,
					IP:               ip.String(),
				})
			}),
			User:  "root",
			Group: "root",
			Perm:  0664,
		},
		world.NewConfiguraionBuilder().WithApplier(&remote.Command{
			SSH:     s.SSH,
			Command: "systemctl restart nginx",
		}),
	}
}

func (s *NginxProxy) Applier() (world.Applier, error) {
	return nil, nil
}
