package docker

import (
	"context"

	"github.com/bborbe/world"
	"github.com/pkg/errors"
)

func UploadIfNeeded(ctx context.Context, uploader world.Uploader) error {
	ok, err := uploader.Satisfied(ctx)
	if err != nil {
		return errors.Wrap(err, "check satisfied failed")
	}
	if !ok {
		if err := uploader.Upload(ctx); err != nil {
			return errors.Wrap(err, "upload failed")
		}
	}
	return nil
}
