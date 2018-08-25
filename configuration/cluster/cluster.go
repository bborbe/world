package cluster

import (
	"context"

	"github.com/bborbe/world/pkg/k8s"
	"github.com/pkg/errors"
)

type Cluster struct {
	Context   k8s.Context
	NfsServer k8s.PodNfsServer
}

func (w Cluster) Validate(ctx context.Context) error {
	if w.Context == "" {
		return errors.New("Context missing")
	}
	if w.NfsServer == "" {
		return errors.New("NfsServer missing")
	}
	return nil
}
