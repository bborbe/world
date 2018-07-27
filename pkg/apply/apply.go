package apply

import (
	"github.com/bborbe/run"
	"github.com/bborbe/world"
	"github.com/golang/glog"
	"golang.org/x/net/context"
)

type Applier struct {
	Deployer []world.Deployer
	Uploader []world.Uploader
	Builder  []world.Builder
}

func (a *Applier) Apply(ctx context.Context) error {
	glog.V(4).Infof("apply app ...")
	defer glog.V(4).Infof("apply app finished")

	var list []run.RunFunc
	for _, builder := range a.Builder {
		list = append(list, func(ctx context.Context) error {
			ok, err := builder.Satisfied(ctx)
			if err != nil {
				return err
			}
			if ok {
				return nil
			}
			return builder.Build(ctx)
		})
	}

	for _, uploader := range a.Uploader {
		list = append(list, func(ctx context.Context) error {
			ok, err := uploader.Satisfied(ctx)
			if err != nil {
				return err
			}
			if ok {
				return nil
			}
			return uploader.Upload(ctx)
		})
	}

	for _, deployer := range a.Deployer {
		list = append(list, func(ctx context.Context) error {
			ok, err := deployer.Satisfied(ctx)
			if err != nil {
				return err
			}
			if ok {
				return nil
			}
			return deployer.Deploy(ctx)
		})
	}

	return run.Sequential(ctx, list...)
}
