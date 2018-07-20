package docker

import (
	"context"

	"github.com/bborbe/run"
	"github.com/bborbe/world"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type BuildAll struct {
}

func (b *BuildAll) Build(ctx context.Context) error {
	glog.V(1).Infof("build all ...")
	var list []run.RunFunc
	for _, app := range world.Apps {
		builder, err := BuilderForApp(app)
		if err != nil {
			return errors.Wrap(err, "get builder for app failed")
		}
		list = append(list, builder.Build)
	}
	glog.V(1).Infof("found %d builds", len(list))
	return errors.Wrap(run.CancelOnFirstError(ctx, list...), "build all failed")
}
