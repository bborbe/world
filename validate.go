package world

import (
	"context"
)

func Validate(ctx context.Context, cfg Configuration) error {
	if err := cfg.Validate(ctx); err != nil {
		return err
	}
	if cfg.Applier() != nil {
		if err := cfg.Applier().Validate(ctx); err != nil {
			return err
		}
	}
	for _, child := range cfg.Childs() {
		if err := Validate(ctx, child); err != nil {
			return err
		}
	}
	return nil
}
