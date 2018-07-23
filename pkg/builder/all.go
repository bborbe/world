package builder

import (
	"context"

	"github.com/bborbe/run"
	"github.com/bborbe/world"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type BuildAll struct {
	Apps world.Apps
}

func (b *BuildAll) Build(ctx context.Context) error {
	glog.V(1).Infof("build all ...")
	var list []run.RunFunc
	for _, app := range b.Apps {
		list = append(list, app.Deployer.GetUploader().GetBuilder().Build)
	}
	glog.V(1).Infof("found %d builds", len(list))
	return errors.Wrap(run.CancelOnFirstError(ctx, list...), "build all failed")
}
