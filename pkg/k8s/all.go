package k8s

import (
	"context"

	"github.com/bborbe/run"
	"github.com/bborbe/world"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type DeployAll struct {
	Apps world.Apps
}

func (b *DeployAll) Deploy(ctx context.Context) error {
	glog.V(1).Infof("deploy all ...")
	var list []run.RunFunc
	for _, app := range b.Apps {
		list = append(list, app.Deployer.Deploy)
	}
	glog.V(1).Infof("found %d deploys", len(list))
	return errors.Wrap(run.CancelOnFirstError(ctx, list...), "deploy all failed")
}
