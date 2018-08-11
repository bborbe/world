package world

import (
	"context"

	"github.com/bborbe/run"
	"github.com/golang/glog"
)

func Apply(ctx context.Context, cfg Configuration) error {
	glog.V(4).Infof("apply configuration ...")
	if cfg.Applier() != nil {
		ok, err := cfg.Applier().Satisfied(ctx)
		if err != nil {
			return err
		}
		if ok {
			glog.V(4).Infof("already satisfied => skip")
			return nil
		}
	}
	glog.V(4).Infof("found %d childs", len(cfg.Childs()))

	var list []run.RunFunc
	for _, child := range cfg.Childs() {
		list = append(list, func(child Configuration) run.RunFunc {
			return func(ctx context.Context) error {
				return Apply(ctx, child)
			}
		}(child))
	}
	if err := run.Sequential(ctx, list...); err != nil {
		return err
	}
	if cfg.Applier() != nil {
		if err := cfg.Applier().Apply(ctx); err != nil {
			return err
		}
	}
	glog.V(4).Infof("apply configuration finished")
	return nil
}
