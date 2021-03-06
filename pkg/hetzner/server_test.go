// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hetzner

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/bborbe/world/pkg/ssh"
)

func TestUserData(t *testing.T) {
	tempFile, err := ioutil.TempFile("", "")
	tempFile.WriteString("hello world")
	tempFile.Close()
	defer func() {
		_ = os.Remove(tempFile.Name())
	}()
	s := Server{
		User:          ssh.User("tester"),
		PublicKeyPath: ssh.PublicKeyPath(tempFile.Name()),
	}
	userdata, err := s.userdata()
	if err != nil {
		t.Fatal(err)
	}
	if userdata != `#cloud-config
users:
- name: tester
  sudo: ALL=(ALL) NOPASSWD:ALL
  ssh_authorized_keys:
  - hello world
  shell: /bin/bash
` {
		t.Fatal(fmt.Sprintf("invalid content %s", userdata))
	}
}
