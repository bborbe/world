// Copyright (c) 2019 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package backup

import "github.com/bborbe/world/configuration/app"

var Netcup = app.BackupTarget{
	User:      "root",
	Host:      "v22016124049440903.goodsrv.de",
	Port:      2222,
	Directory: "/data/",
	Excludes: []string{
		"/cdrom",
		"/data/confluence-data/analytics-logs",
		"/data/confluence-data/backups",
		"/data/confluence-data/plugin-cache",
		"/data/confluence-data/plugins-cache",
		"/data/confluence-data/plugins-osgi-cache",
		"/data/confluence-data/plugins-temp",
		"/data/confluence-data/temp",
		"/data/confluence-data/webresource-temp",
		"/data/jenkins-slave-docker/workspace",
		"/data/jenkins-slave-golang/workspace",
		"/data/jenkins-slave-java/.m2",
		"/data/jenkins-slave-java/workspace",
		"/dev",
		"/media",
		"/proc",
		"/rsync",
		"/run",
		"/sys",
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
