// Copyright (c) 2020 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service

import (
	"context"
	"fmt"

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
	SSH          *ssh.SSH
	Domain       network.Host
	Target       string
	Requirements []world.Configuration
}

func (s *NginxProxy) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		s.SSH,
		s.Domain,
	)
}

func (s *NginxProxy) Children() []world.Configuration {
	var result []world.Configuration
	result = append(result, s.Requirements...)
	result = append(result, s.proxy()...)
	return result
}

func (s *NginxProxy) proxy() []world.Configuration {
	return []world.Configuration{
		&remote.File{
			SSH:  s.SSH,
			Path: file.Path(fmt.Sprintf("/etc/nginx/sites-enabled/%s.conf", s.Domain)),
			Content: content.Func(func(ctx context.Context) ([]byte, error) {
				return template.Render(`
server {
	server_name {{ .Domain }};

	location / {
		proxy_set_header X-Forwarded-Host $host:$server_port;
		proxy_set_header X-Forwarded-Server $host;
		proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
		proxy_set_header Host $host;
		proxy_pass {{ .Target }}/;
	}

	listen 443 ssl;
	ssl_certificate /etc/letsencrypt/live/{{ .Domain }}/fullchain.pem;
	ssl_certificate_key /etc/letsencrypt/live/{{ .Domain }}/privkey.pem;
	include /etc/letsencrypt/options-ssl-nginx.conf;
	ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem;
}

server {
	listen 80;

	server_name {{ .Domain }};
	return 301 https://$host$request_uri;
}
`, struct {
					Domain string
					Target string
				}{
					Domain: s.Domain.String(),
					Target: s.Target,
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
