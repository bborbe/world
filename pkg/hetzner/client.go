// Copyright (c) 2019 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hetzner

import (
	"context"
	"net"

	"github.com/bborbe/world/pkg/k8s"
	"github.com/pkg/errors"
)

type Client interface {
	GetIP(ctx context.Context, key ApiKey, name k8s.Context) (net.IP, error)
}

func NewClient() Client {
	return &client{}
}

type client struct {
}

func (c *client) GetIP(ctx context.Context, key ApiKey, name k8s.Context) (net.IP, error) {
	server, _, err := key.Client().Server.GetByName(ctx, name.String())
	if err != nil {
		return nil, errors.Wrap(err, "get server failed")
	}
	return server.PublicNet.IPv4.IP, nil
}

type clientDummy struct {
}

func (c *clientDummy) GetIP(ctx context.Context, key ApiKey, name k8s.Context) (net.IP, error) {
	return net.ParseIP("1.2.3.4"), nil
}

func NewCLientDummy() Client {
	return &clientDummy{}
}
