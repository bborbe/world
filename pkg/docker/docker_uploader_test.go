// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package docker

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestUploaderCallsUrl(t *testing.T) {
	called := false
	var calledUrl string
	getUrl = func(url string) (resp *http.Response, err error) {
		called = true
		calledUrl = url
		return &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(&bytes.Buffer{}),
		}, nil
	}
	u := Uploader{
		Image: Image{
			Repository: "bborbe/poste.io",
			Tag:        "1.0.6",
		},
	}
	_, _ = u.Satisfied(context.Background())
	if !called {
		t.Fatal("get not called")
	}
	if calledUrl != "https://hub.docker.com/v2/repositories/bborbe/poste.io/tags/1.0.6/" {
		t.Fatalf("unexpected url: %s", calledUrl)
	}
}
