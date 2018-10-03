package remote

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

type Path string

func (p Path) String() string {
	return string(p)
}

func (f Path) Validate(ctx context.Context) error {
	if f == "" {
		return errors.New("Path missing")
	}
	return nil
}

type HasContent interface {
	Content(ctx context.Context) ([]byte, error)
}

type StaticContent []byte

func (s StaticContent) Content(ctx context.Context) ([]byte, error) {
	return s, nil
}

type ContentFunc func(ctx context.Context) ([]byte, error)

func (s ContentFunc) Content(ctx context.Context) ([]byte, error) {
	return s(ctx)
}

type User string

func (f User) Validate(ctx context.Context) error {
	if f == "" {
		return errors.New("User missing")
	}
	return nil
}

type Group string

func (f Group) Validate(ctx context.Context) error {
	if f == "" {
		return errors.New("Group missing")
	}
	return nil
}

type Perm uint32

func (f Perm) Validate(ctx context.Context) error {
	if f == 0 {
		return errors.New("Perm missing")
	}
	return nil
}

func (f Perm) String() string {
	return fmt.Sprintf("%04o", uint32(f))
}
