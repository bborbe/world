// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package openvpn

import (
	"context"

	"github.com/bborbe/world/configuration/service"
	"github.com/bborbe/world/pkg/apt"
	"github.com/bborbe/world/pkg/file"
	"github.com/bborbe/world/pkg/network"
	"github.com/bborbe/world/pkg/remote"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Server struct {
	SSH         *ssh.SSH
	ServerName  ServerName
	ServerIPNet network.IPNet
	Routes      Routes
}

func (s *Server) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		s.SSH,
		s.ServerName,
		s.ServerIPNet,
		s.Routes,
	)
}

func (s *Server) Children() []world.Configuration {
	serverConfig := &ServerConfig{
		ServerName:  s.ServerName,
		ServerIPNet: s.ServerIPNet,
		Routes:      s.Routes,
	}
	return []world.Configuration{
		&remote.File{
			SSH:     s.SSH,
			Path:    file.Path("/etc/default/openvpn"),
			User:    "root",
			Group:   "root",
			Perm:    0644,
			Content: serverConfig.OpenvpnDefaultConf(),
		},
		&service.Directory{
			SSH:   s.SSH,
			Path:  file.Path("/etc/openvpn/keys"),
			User:  "root",
			Group: "root",
			Perm:  0700,
		},
		&remote.File{
			SSH:     s.SSH,
			Path:    file.Path("/etc/openvpn/server.conf"),
			User:    "root",
			Group:   "root",
			Perm:    0600,
			Content: serverConfig.ServerConfigContent(),
		},
		&remote.FileLocalCached{
			SSH:       s.SSH,
			Path:      file.Path("/etc/openvpn/keys/ta.key"),
			LocalPath: serverConfig.LocalPathTaKey(),
			User:      "root",
			Group:     "root",
			Perm:      0600,
			Content:   serverConfig.TAKey(),
		},
		&remote.FileLocalCached{
			SSH:       s.SSH,
			Path:      file.Path("/etc/openvpn/keys/dh.pem"),
			LocalPath: serverConfig.LocalPathDhPem(),
			User:      "root",
			Group:     "root",
			Perm:      0600,
			Content:   serverConfig.DHPem(),
		},
		&remote.FileLocalCached{
			SSH:       s.SSH,
			Path:      file.Path("/etc/openvpn/keys/ca.key"),
			LocalPath: serverConfig.LocalPathCAPrivateKey(),
			User:      "root",
			Group:     "root",
			Perm:      0600,
			Content:   serverConfig.CAPrivateKey(),
		},
		&remote.FileLocalCached{
			SSH:       s.SSH,
			Path:      file.Path("/etc/openvpn/keys/ca.crt"),
			LocalPath: serverConfig.LocalPathCaCrt(),
			User:      "root",
			Group:     "root",
			Perm:      0600,
			Content:   serverConfig.CaCrt(),
		},
		&remote.FileLocalCached{
			SSH:       s.SSH,
			Path:      file.Path("/etc/openvpn/keys/server.crt"),
			LocalPath: serverConfig.LocalPathServerCrt(),
			User:      "root",
			Group:     "root",
			Perm:      0600,
			Content:   serverConfig.ServerCrt(),
		},
		&remote.FileLocalCached{
			SSH:       s.SSH,
			Path:      file.Path("/etc/openvpn/keys/server.key"),
			LocalPath: serverConfig.LocalPathServerKey(),
			User:      "root",
			Group:     "root",
			Perm:      0600,
			Content:   serverConfig.ServerKey(),
		},
		world.NewConfiguraionBuilder().WithApplier(&remote.Iptables{
			SSH:  s.SSH,
			Port: 563,
		}),
		world.NewConfiguraionBuilder().WithApplier(&apt.Update{
			SSH: s.SSH,
		}),
		world.NewConfiguraionBuilder().WithApplier(&apt.Install{
			SSH:     s.SSH,
			Package: "openvpn",
		}),
		world.NewConfiguraionBuilder().WithApplier(&apt.Autoremove{
			SSH: s.SSH,
		}),
		world.NewConfiguraionBuilder().WithApplier(&apt.Clean{
			SSH: s.SSH,
		}),
		world.NewConfiguraionBuilder().WithApplier(&remote.ServiceStart{
			SSH:  s.SSH,
			Name: "openvpn",
		}),
	}
}

func (s *Server) Applier() (world.Applier, error) {
	return nil, nil
}
