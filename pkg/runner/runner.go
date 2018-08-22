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
	Name    string
	Applier world.Applier
	Runners []Runner
}

func New(configuration world.Configuration) (*Runner, error) {
	applier, err := configuration.Applier()
	if err != nil {
		return nil, err
	}
	var runners []Runner
	for _, configuration := range configuration.Children() {
		runner, err := New(configuration)
		if err != nil {
			return nil, err
		}
		runners = append(runners, *runner)
	}
	return &Runner{
		Applier: applier,
		Runners: runners,
		Name:    reflect.TypeOf(configuration).String(),
	}, nil
}

func (r Runner) Apply(ctx context.Context) error {
	return apply(ctx, r, nil)
}

func (r Runner) Validate(ctx context.Context) error {
	return validate(ctx, r, nil)
}

func apply(ctx context.Context, cfg Runner, path []string) error {
	path = append(path, cfg.Name)
	glog.V(4).Infof("apply configuration %s ...", strings.Join(path, " -> "))
	if cfg.Applier != nil {
		ok, err := cfg.Applier.Satisfied(ctx)
		if err != nil {
			return errors.Wrap(err, "check satisfied failed")
		}
		if ok {
			glog.V(4).Infof("already satisfied => skip")
			return nil
		}
	}
	glog.V(4).Infof("found %d children", len(cfg.Runners))

	var list []run.RunFunc
	for _, child := range cfg.Runners {
		list = append(list, func(child Runner) run.RunFunc {
			return func(ctx context.Context) error {
				return apply(ctx, child, path)
			}
		}(child))
	}
	if err := run.Sequential(ctx, list...); err != nil {
		return errors.Wrap(err, "apply children failed")
	}
	if cfg.Applier != nil {
		if err := cfg.Applier.Apply(ctx); err != nil {
			return errors.Wrap(err, "apply failed")
		}
	}
	glog.V(2).Infof("configuration %s applied", strings.Join(path, " -> "))
	return nil
}

func validate(ctx context.Context, cfg Runner, path []string) error {
	path = append(path, cfg.Name)
	glog.V(4).Infof("validate configuration %s", strings.Join(path, " -> "))
	if cfg.Applier != nil {
		if err := cfg.Applier.Validate(ctx); err != nil {
			return errors.Wrapf(err, "in %s %s", strings.Join(path, " -> "), cfg.Name)
		}
	}
	for _, child := range cfg.Runners {
		if err := validate(ctx, child, path); err != nil {
			return err
		}
	}
	glog.V(2).Infof("configuration %s is valid", strings.Join(path, " -> "))
	return nil
}
