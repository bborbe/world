package apply

import (
	"context"

	"github.com/bborbe/run"
	"github.com/bborbe/world"
	"github.com/golang/glog"
)

type ApplyAll struct {
	Apps world.Apps
}

func (a *ApplyAll) Apply(ctx context.Context) error {
	var list []run.RunFunc
	for _, app := range a.Apps {
		list = append(list, buildFunc(app))
	}
	return run.All(ctx, list...)
}

func buildFunc(app world.App) run.RunFunc {
	return func(ctx context.Context) error {
		glog.V(4).Infof("apply app %s ...", app.Name.String())
		defer glog.V(4).Infof("apply app %s finished", app.Name.String())
		return run.Sequential(
			ctx,
			app.Validate,
			app.Deployer.Deploy,
		)
	}
}
