// Copyright (c) 2019 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package openvpn

import (
	"context"

	"github.com/bborbe/world/pkg/local"
	"github.com/bborbe/world/pkg/world"
)

type LocalClient struct {
	ClientName    ClientName
	ServerName    ServerName
	ServerAddress ServerAddress
}

func (l *LocalClient) Children() []world.Configuration {
	clientConfig := l.clientConfig()
	return []world.Configuration{
		world.NewConfiguraionBuilder().WithApplier(
			&local.FileContent{
				Path:    clientConfig.LocalPathConfig(),
				Content: clientConfig.ConfigContent(),
			},
		),
		world.NewConfiguraionBuilder().WithApplier(
			&local.FileContent{
				Path:    clientConfig.LocalPathCaCrt(),
				Content: clientConfig.CaCrt(),
			},
		),
		world.NewConfiguraionBuilder().WithApplier(
			&local.FileContent{
				Path:    clientConfig.LocalPathTaKey(),
				Content: clientConfig.TAKey(),
			},
		),
		world.NewConfiguraionBuilder().WithApplier(
			&local.FileContent{
				Path:    clientConfig.LocalPathClientKey(),
				Content: clientConfig.ClientKey(),
			},
		),
		world.NewConfiguraionBuilder().WithApplier(
			&local.FileContent{
				Path:    clientConfig.LocalPathClientCrt(),
				Content: clientConfig.ClientCrt(),
			},
		),
	}
}

func (l *LocalClient) Applier() (world.Applier, error) {
	return nil, nil
}

func (l *LocalClient) Validate(ctx context.Context) error {
	return l.clientConfig().Validate(ctx)
}

func (l *LocalClient) clientConfig() ClientConfig {
	return ClientConfig{
		ClientName:    l.ClientName,
		ServerAddress: l.ServerAddress,
		ServerConfig: ServerConfig{
			ServerName: l.ServerName,
		},
	}
}
