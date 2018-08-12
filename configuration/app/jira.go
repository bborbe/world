package app

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type Jira struct {
	Cluster cluster.Cluster
}

func (j *Jira) Childs() []world.Configuration {
	return []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   j.Cluster.Context,
			Namespace: "jira",
		},
	}
}

func (j *Jira) Applier() world.Applier {
	return nil
}

func (j *Jira) Validate(ctx context.Context) error {
	glog.V(4).Infof("validate jira app ...")
	if err := j.Cluster.Validate(ctx); err != nil {
		return errors.Wrap(err, "validate jira app failed")
	}
	return nil
}
