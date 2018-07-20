package k8s

import (
	"context"

	"github.com/bborbe/run"
	"github.com/bborbe/world"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type DeployAll struct {
}

func (b *DeployAll) Deploy(ctx context.Context) error {
	glog.V(1).Infof("deploy all ...")
	var list []run.RunFunc
	for _, app := range world.Apps {
		deployer, err := DeployerForApp(app)
		if err != nil {
			return errors.Wrap(err, "get deployer for app failed")
		}
		list = append(list, deployer.Deploy)
	}
	glog.V(1).Infof("found %d deploys", len(list))
	return errors.Wrap(run.CancelOnFirstError(ctx, list...), "deploy all failed")
}
