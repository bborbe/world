package app

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type Jenkins struct {
	Cluster cluster.Cluster
}

func (j *Jenkins) Children() []world.Configuration {
	return []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   j.Cluster.Context,
			Namespace: "jenkins",
		},
	}
}

func (j *Jenkins) Applier() world.Applier {
	return nil
}

func (j *Jenkins) Validate(ctx context.Context) error {
	glog.V(4).Infof("validate jenkins app ...")
	if err := j.Cluster.Validate(ctx); err != nil {
		return errors.Wrap(err, "validate jenkins app failed")
	}
	return nil
}
