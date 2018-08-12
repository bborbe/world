package world

import (
	"context"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

func Validate(ctx context.Context, cfg Configuration) error {
	return validate(ctx, cfg, nil)
}

func validate(ctx context.Context, cfg Configuration, path []string) error {
	path = append(path, nameOf(cfg))
	if err := cfg.Validate(ctx); err != nil {
		return errors.Wrapf(err, "in %s", strings.Join(path, " -> "))
	}
	if cfg.Applier() != nil {
		if err := cfg.Applier().Validate(ctx); err != nil {
			return errors.Wrapf(err, "in %s %s", strings.Join(path, " -> "), nameOf(cfg.Applier()))
		}
	}
	for _, child := range cfg.Childs() {
		if err := validate(ctx, child, path); err != nil {
			return err
		}
	}
	return nil
}

func nameOf(obj interface{}) string {
	of := reflect.TypeOf(obj)
	return of.String()
}
