// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

var getUrl = http.Get

type Uploader struct {
	Image Image
}

func (u *Uploader) Apply(ctx context.Context) error {
	glog.V(1).Infof("upload docker image %s ...", u.Image.String())
	cmd := exec.CommandContext(ctx, "docker", "push", u.Image.String())
	if glog.V(4) {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "upload docker image failed")
	}
	glog.V(1).Infof("upload docker image %s finished", u.Image.String())
	return nil
}

func (u *Uploader) Satisfied(ctx context.Context) (bool, error) {
	glog.V(3).Infof("check image %s exists an registry ...", u.Image.String())
	url := fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/tags/%s/", u.Image.Repository, u.Image.Tag)
	resp, err := getUrl(url)
	if err != nil {
		return false, errors.Wrapf(err, "get %s failed", url)
	}
	if resp.StatusCode/100 != 2 {
		glog.V(1).Infof("image %s is missing on registry", u.Image.String())
		return false, nil
	}
	defer resp.Body.Close()
	var data struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		glog.V(1).Infof("decode json failed: %v", err)
		return false, errors.Wrap(err, "decode json failed")
	}
	if data.Name != u.Image.Tag.String() {
		glog.V(2).Infof("tag mismatch")
		return false, nil
	}
	glog.V(3).Infof("image %s exists on registry", u.Image.String())
	return true, nil
}

func (u *Uploader) Validate(ctx context.Context) error {
	glog.V(4).Infof("validate docker uploader ...")
	if err := u.Image.Validate(ctx); err != nil {
		return errors.Wrap(err, "validate golang uploader failed")
	}
	return nil
}
