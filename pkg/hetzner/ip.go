// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hetzner

import (
	"context"
	"net"

	"github.com/golang/glog"
	"github.com/pkg/errors"

	"github.com/bborbe/world/pkg/deployer"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
)

type IP struct {
	Client Client
	ApiKey deployer.SecretValue
	Name   k8s.Context
}

func (i IP) Validate(ctx context.Context) error {
	if i.Client == nil {
		return errors.New("client missing")
	}
	return validation.Validate(ctx,
		i.Name,
		i.ApiKey,
	)
}

func (i IP) IP(ctx context.Context) (net.IP, error) {
	apiKey, err := i.ApiKey.Value(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get apikey failed")
	}
	glog.V(4).Infof("get ip for %s with api key %s", i.Name.String(), string(apiKey))
	return i.Client.GetIP(ctx, ApiKey(apiKey), i.Name)
}
