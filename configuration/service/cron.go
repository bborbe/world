// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/pkg/errors"

	"github.com/bborbe/world/pkg/content"
	"github.com/bborbe/world/pkg/file"
	"github.com/bborbe/world/pkg/remote"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

var cronNameValidChars = regexp.MustCompile(`^[a-z0-9_-]+$`)
var cronNameRemoveInvalidChars = regexp.MustCompile(`[^a-z0-9_-]`)
var cronNameUnderscore = regexp.MustCompile(`_+`)

func BuildCronName(parts ...string) CronName {
	result := strings.Join(parts, "_")
	result = strings.ToLower(result)
	result = cronNameRemoveInvalidChars.ReplaceAllString(result, `_`)
	result = cronNameUnderscore.ReplaceAllString(result, `_`)
	return CronName(result)
}

type CronName string

func (c CronName) Validate(ctx context.Context) error {
	if !cronNameValidChars.MatchString(c.String()) {
		return errors.Errorf("cron name invalid")
	}
	return nil
}

func (c CronName) String() string {
	return string(c)
}

type CronPath string

func (c CronPath) Validate(ctx context.Context) error {
	return nil
}

func (c CronPath) String() string {
	if c == "" {
		return "/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin"
	}
	return string(c)
}

type CronShell string

func (c CronShell) Validate(ctx context.Context) error {
	return nil
}

func (c CronShell) String() string {
	if c == "" {
		return "/bin/sh"
	}
	return string(c)
}

type CronUser string

func (c CronUser) Validate(ctx context.Context) error {
	return nil
}

func (c CronUser) String() string {
	if c == "" {
		return "root"
	}
	return string(c)
}

type CronExpression string

func (c CronExpression) Validate(ctx context.Context) error {
	if c == "" {
		return errors.Errorf("cron expression empty")
	}
	return nil
}

func (c CronExpression) String() string {
	return string(c)
}

type CronSchedule string

func (c CronSchedule) Validate(ctx context.Context) error {
	parts := strings.FieldsFunc(c.String(), func(r rune) bool {
		return r == ' '
	})
	if len(parts) != 5 {
		return errors.Errorf("cron schedule invalid")
	}
	return nil
}

func (c CronSchedule) String() string {
	return string(c)
}

type Cron struct {
	SSH        *ssh.SSH
	Expression content.HasContent
	Name       CronName
	Path       CronPath
	Schedule   CronSchedule
	Shell      CronShell
	User       CronUser
}

func (d *Cron) Children(ctx context.Context) (world.Configurations, error) {
	return world.Configurations{
		&remote.File{
			SSH:     d.SSH,
			Path:    d.path(),
			Content: d.content(),
			User:    "root",
			Group:   "root",
			Perm:    0644,
		},
	}, nil
}

func (d *Cron) Applier() (world.Applier, error) {
	return nil, nil
}

func (d *Cron) Validate(ctx context.Context) error {
	if d.Expression == nil {
		return errors.Errorf("expression missing")
	}
	return validation.Validate(
		ctx,
		d.SSH,
		d.Name,
		d.Path,
		d.Schedule,
		d.Shell,
		d.User,
	)
}

func (d *Cron) path() file.HasPath {
	return file.Path(fmt.Sprintf("/etc/cron.d/%s", d.Name))
}

func (d *Cron) content() content.HasContent {
	return content.Func(func(ctx context.Context) ([]byte, error) {
		expression, err := d.Expression.Content(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "get expression failed")
		}

		buf := &bytes.Buffer{}
		fmt.Fprintf(buf, "SHELL=%s\n", d.Shell)
		fmt.Fprintf(buf, "PATH=%s\n", d.Path)
		fmt.Fprintln(buf)
		fmt.Fprintf(buf, "%s %s %s\n", d.Schedule, d.User, string(expression))
		return buf.Bytes(), nil
	})
}
