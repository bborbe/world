package world

import (
	"context"

	"strings"

	"github.com/bborbe/run"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

//go:generate counterfeiter -o mocks/applier.go --fake-name Applier . Applier
type Applier interface {
	Satisfied(ctx context.Context) (bool, error)
	Apply(ctx context.Context) error
	Validate(ctx context.Context) error
}

func Apply(ctx context.Context, cfg Configuration) error {
	return apply(ctx, cfg, nil)
}

func apply(ctx context.Context, cfg Configuration, path []string) error {
	path = append(path, nameOf(cfg))
	glog.V(4).Infof("apply configuration %s ...", strings.Join(path, " -> "))
	if cfg.Applier() != nil {
		ok, err := cfg.Applier().Satisfied(ctx)
		if err != nil {
			return errors.Wrap(err, "check satisfied failed")
		}
		if ok {
			glog.V(4).Infof("already satisfied => skip")
			return nil
		}
	}
	glog.V(4).Infof("found %d children", len(cfg.Children()))

	var list []run.RunFunc
	for _, child := range cfg.Children() {
		list = append(list, func(child Configuration) run.RunFunc {
			return func(ctx context.Context) error {
				return apply(ctx, child, path)
			}
		}(child))
	}
	if err := run.Sequential(ctx, list...); err != nil {
		return errors.Wrap(err, "apply children failed")
	}
	if cfg.Applier() != nil {
		if err := cfg.Applier().Apply(ctx); err != nil {
			return errors.Wrap(err, "apply failed")
		}
	}
	glog.V(4).Infof("apply configuration %s finished", strings.Join(path, " -> "))
	return nil
}
