package cluster

import (
	"context"

	"github.com/bborbe/world"
	"github.com/pkg/errors"
)

type Cluster struct {
	Context   world.Context
	NfsServer world.MountNfsServer
}

func (c *Cluster) Validate(ctx context.Context) error {
	if c.Context == "" {
		return errors.New("cluster context missing")
	}
	if c.NfsServer == "" {
		return errors.New("cluster nfs-server missing")
	}
	return nil
}
