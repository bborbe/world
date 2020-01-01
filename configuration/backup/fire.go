// Copyright (c) 2019 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package backup

import "github.com/bborbe/world/configuration/app"

var Fire = app.BackupTarget{
	User:      "root",
	Host:      "fire.hm.benjamin-borbe.de",
	Port:      22,
	Directory: "/",
	Excludes: []string{
		"/backup",
		"/cdrom",
		"/dev",
		"/media",
		"/proc",
		"/rsync",
		"/run",
		"/sys",
		"/timemachine",
		"/tmp",
		"/var/backup",
		"/var/cache",
		"/var/cache/apt/archives",
		"/var/lib/docker",
		"/var/lib/kubelet",
		"/var/lib/lightdm/.gvfs",
		"/var/lib/lxcfs",
		"/var/lock",
		"/var/log",
		"/var/run",
		"/var/tmp",
	},
}
