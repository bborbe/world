package runner

import (
	"context"
	"reflect"
	"strings"

	"github.com/bborbe/run"
	"github.com/bborbe/world"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type Runner struct {
	Configuration world.Configuration
}

func (r *Runner) Apply(ctx context.Context) error {
	return apply(ctx, r.Configuration, nil)
}

func (r *Runner) Validate(ctx context.Context) error {
	return validate(ctx, r.Configuration, nil)
}

func apply(ctx context.Context, cfg world.Configuration, path []string) error {
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
		list = append(list, func(child world.Configuration) run.RunFunc {
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

func validate(ctx context.Context, cfg world.Configuration, path []string) error {
	path = append(path, nameOf(cfg))
	glog.V(4).Infof("validate configuration %s ...", strings.Join(path, " -> "))
	if err := cfg.Validate(ctx); err != nil {
		return errors.Wrapf(err, "in %s", strings.Join(path, " -> "))
	}
	if cfg.Applier() != nil {
		if err := cfg.Applier().Validate(ctx); err != nil {
			return errors.Wrapf(err, "in %s %s", strings.Join(path, " -> "), nameOf(cfg.Applier()))
		}
	}
	for _, child := range cfg.Children() {
		if err := validate(ctx, child, path); err != nil {
			return err
		}
	}
	glog.V(4).Infof("validate configuration %s finished", strings.Join(path, " -> "))
	return nil
}

func nameOf(obj interface{}) string {
	return reflect.TypeOf(obj).String()
}
