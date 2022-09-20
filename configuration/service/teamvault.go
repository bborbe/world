// Copyright (c) 2020 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/pkg/content"
	"github.com/bborbe/world/pkg/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/file"
	"github.com/bborbe/world/pkg/network"
	"github.com/bborbe/world/pkg/remote"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Teamvault struct {
	SSH              *ssh.SSH
	AppPort          network.Port
	DBPort           network.Port
	Domain           string
	DatabasePassword deployer.SecretValue
	SmtpPassword     deployer.SecretValue
	SmtpUsername     deployer.SecretValue
	LdapPassword     deployer.SecretValue
	SecretKey        deployer.SecretValue
	FernetKey        deployer.SecretValue
	Salt             deployer.SecretValue
	Requirements     []world.Configuration
}

func (t *Teamvault) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.SSH,
		t.AppPort,
		t.DBPort,
	)
}

func (t *Teamvault) Children(ctx context.Context) (world.Configurations, error) {
	var result []world.Configuration
	result = append(result, t.Requirements...)
	result = append(result, &Postgres{
		SSH:                  t.SSH,
		Name:                 "teamvault-postgres",
		PostgresVersion:      "10.5",
		PostgresDatabaseName: "teamvault",
		PostgresInitDbArgs:   "--encoding=UTF8 --lc-collate=en_US.UTF-8 --lc-ctype=en_US.UTF-8 -T template0",
		PostgresUsername:     deployer.SecretValueStatic("teamvault"),
		PostgresPassword:     t.DatabasePassword,
		Port:                 t.DBPort,
		DataPath:             "/home/teamvault-postgres",
	})
	result = append(result, t.teamvault()...)
	return result, nil
}

func (t *Teamvault) teamvault() []world.Configuration {
	version := "0.9.2"
	image := docker.Image{
		Repository: "bborbe/teamvault",
		Tag:        docker.TagWithTime(version, time.Now()),
	}
	envFile := "/home/teamvault.environment"
	return world.Configurations{
		&build.Teamvault{
			Image:   image,
			Version: version,
		},
		&remote.File{
			SSH:  t.SSH,
			Path: file.Path(envFile),
			Content: content.Func(func(ctx context.Context) ([]byte, error) {
				databasePassword, err := t.DatabasePassword.Value(ctx)
				if err != nil {
					return nil, err
				}
				secretKey, err := t.SecretKey.Value(ctx)
				if err != nil {
					return nil, err
				}
				fernetKey, err := t.FernetKey.Value(ctx)
				if err != nil {
					return nil, err
				}
				salt, err := t.Salt.Value(ctx)
				if err != nil {
					return nil, err
				}
				ldapPassword, err := t.LdapPassword.Value(ctx)
				if err != nil {
					return nil, err
				}
				dbPort, err := t.DBPort.Port(ctx)
				if err != nil {
					return nil, err
				}
				return EnvFile{
					"BASE_URL":                 fmt.Sprintf("https://%s", t.Domain),
					"DEBUG":                    "disabled",
					"EMAIL_ENABLED":            "true",
					"EMAIL_HOST":               "localhost",
					"EMAIL_PORT":               "25",
					"EMAIL_USER":               "",
					"EMAIL_PASSWORD":           "",
					"EMAIL_USE_TLS":            "False",
					"EMAIL_USE_SSL":            "False",
					"DATABASE_HOST":            "localhost",
					"DATABASE_PORT":            strconv.Itoa(dbPort),
					"DATABASE_NAME":            "teamvault",
					"DATABASE_USER":            "teamvault",
					"DATABASE_PASSWORD":        string(databasePassword),
					"SECRET_KEY":               string(secretKey),
					"FERNET_KEY":               string(fernetKey),
					"SALT":                     string(salt),
					"LDAP_ENABLED":             "true",
					"LDAP_SERVER_URI":          "ldap://localhost",
					"LDAP_BIND_DN":             "cn=root,dc=benjamin-borbe,dc=de",
					"LDAP_PASSWORD":            string(ldapPassword),
					"LDAP_USER_BASE_DN":        "ou=users,dc=benjamin-borbe,dc=de",
					"LDAP_USER_SEARCH_FILTER":  "(uid=%%(user)s)",
					"LDAP_GROUP_BASE_DN":       "ou=groups,dc=benjamin-borbe,dc=de",
					"LDAP_GROUP_SEARCH_FILTER": "(objectClass=groupOfNames)",
					"LDAP_REQUIRE_GROUP":       "ou=employees,ou=groups,dc=benjamin-borbe,dc=de",
					"LDAP_ADMIN_GROUP":         "ou=admins,ou=groups,dc=benjamin-borbe,dc=de",
					"LDAP_ATTR_EMAIL":          "mail",
					"LDAP_ATTR_FIRST_NAME":     "givenName",
					"LDAP_ATTR_LAST_NAME":      "sn",
					"LDAP_CACHE_TIMEOUT":       "60",
					"SECURE_SSL_REDIRECT":      "true",
				}.Content(ctx)
			}),
			User:  "root",
			Group: "root",
			Perm:  0644,
		},
		&Docker{
			SSH:  t.SSH,
			Name: "teamvault",
			BuildDockerServiceContent: func(ctx context.Context) (*DockerServiceContent, error) {
				return &DockerServiceContent{
					Name:    "teamvault",
					Memory:  400,
					Image:   image,
					HostNet: true,
					EnvironmentFiles: []string{
						envFile,
					},
					Requires: []remote.ServiceName{
						"docker.service",
					},
					After: []remote.ServiceName{
						"docker.service",
					},
					TimeoutStartSec: "20s",
					TimeoutStopSec:  "20s",
				}, nil
			},
		},
		world.NewConfiguraionBuilder().WithApplier(&remote.IptablesAllowInput{
			SSH:      t.SSH,
			Port:     t.AppPort,
			Protocol: network.TCP,
		}),
	}
}

func (t *Teamvault) Applier() (world.Applier, error) {
	return nil, nil
}
