// Copyright (c) 2019 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package openvpn

import (
	"context"

	"github.com/bborbe/world/pkg/local"
	"github.com/bborbe/world/pkg/world"
)

type Client struct {
	ClientName ClientName
	ServerName ServerName
}

func (c *Client) Children() []world.Configuration {
	clientConfig := c.clientConfig()
	return []world.Configuration{
		world.NewConfiguraionBuilder().WithApplier(
			&local.FileContent{
				Path:    clientConfig.LocalPath("config.conf"),
				Content: clientConfig.ConfigContent(),
			},
		),
	}
}

func (c *Client) Applier() (world.Applier, error) {
	return nil, nil
}

func (c *Client) Validate(ctx context.Context) error {
	return c.clientConfig().Validate(ctx)
}

func (c *Client) clientConfig() ClientConfig {
	return ClientConfig{
		ClientName: c.ClientName,
		ServerConfig: ServerConfig{
			ServerName: c.ServerName,
		},
	}
}
