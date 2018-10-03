package hetzner

import (
	"context"
	"net"

	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
)

type IP struct {
	ApiKey deployer.SecretValue
	Name   k8s.Context
}

func (i IP) Validate(ctx context.Context) error {
	return validation.Validate(ctx,
		i.Name,
		i.ApiKey,
	)
}

func (i IP) IP(ctx context.Context) (net.IP, error) {
	bytes, err := i.ApiKey.Value()
	if err != nil {
		return nil, err
	}
	server, _, err := ApiKey(bytes).Client().Server.GetByName(ctx, i.Name.String())
	if err != nil {
		return nil, err
	}
	return server.PublicNet.IPv4.IP, nil

}
