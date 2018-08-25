package app

import (
	"fmt"

	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/component"
	"github.com/bborbe/world/configuration/container"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/pkg/errors"
)

type Teamvault struct {
	Cluster          cluster.Cluster
	Domains          k8s.IngressHosts
	DatabasePassword deployer.SecretValue
	SmtpPassword     deployer.SecretValue
	SmtpUsername     deployer.SecretValue
	LdapPassword     deployer.SecretValue
	SecretKey        deployer.SecretValue
	FernetKey        deployer.SecretValue
	Salt             deployer.SecretValue
}

func (t *Teamvault) Validate(ctx context.Context) error {
	if len(t.Domains) != 1 {
		return errors.New("need 1 domain")
	}
	return validation.Validate(
		ctx,
		t.Cluster,
		t.Domains,
		t.DatabasePassword,
		t.SmtpPassword,
		t.SmtpUsername,
		t.LdapPassword,
		t.SecretKey,
		t.FernetKey,
		t.Salt,
	)
}

func (t *Teamvault) Children() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/teamvault",
		Tag:        "0.7.3",
	}
	ports := []deployer.Port{
		{
			Port:     8000,
			Protocol: "TCP",
			Name:     "http",
		},
	}
	return []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   t.Cluster.Context,
			Namespace: "teamvault",
		},
		&component.Postgres{
			Context:              t.Cluster.Context,
			Namespace:            "teamvault",
			DataNfsPath:          "/data/teamvault-postgres",
			DataNfsServer:        t.Cluster.NfsServer,
			BackupNfsPath:        "/data/teamvault-postgres-backup",
			BackupNfsServer:      t.Cluster.NfsServer,
			PostgresVersion:      "10.5",
			PostgresInitDbArgs:   "--encoding=UTF8 --lc-collate=en_US.UTF-8 --lc-ctype=en_US.UTF-8 -T template0",
			PostgresDatabaseName: "teamvault",
			PostgresUsername: &deployer.SecretValueStatic{
				Content: []byte("teamvault"),
			},
			PostgresPassword: t.DatabasePassword,
		},
		&deployer.SecretDeployer{
			Context:   t.Cluster.Context,
			Namespace: "teamvault",
			Name:      "teamvault",
			Secrets: deployer.Secrets{
				"database-password": t.DatabasePassword,
				"secret-key":        t.SecretKey,
				"fernet-key":        t.FernetKey,
				"salt":              t.Salt,
				"ldap-password":     t.LdapPassword,
			},
		},
		&deployer.DeploymentDeployer{
			Context:      t.Cluster.Context,
			Namespace:    "teamvault",
			Name:         "teamvault",
			Requirements: t.smtp().Requirements(),
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:  "teamvault",
					Image: image,
					Requirement: &build.Teamvault{
						Image: image,
					},
					Ports: ports,
					Resources: k8s.PodResources{
						Limits: k8s.Resources{
							Cpu:    "2000m",
							Memory: "400Mi",
						},
						Requests: k8s.Resources{
							Cpu:    "10m",
							Memory: "100Mi",
						},
					},
					Env: []k8s.Env{
						{
							Name:  "BASE_URL",
							Value: fmt.Sprintf("https://%s", t.Domains[0]),
						},
						{
							Name:  "DEBUG",
							Value: "disabled",
						},
						{
							Name:  "EMAIL_ENABLED",
							Value: "true",
						},
						{
							Name:  "EMAIL_HOST",
							Value: "localhost",
						},
						{
							Name:  "EMAIL_PORT",
							Value: "25",
						},
						{
							Name:  "EMAIL_USER",
							Value: "",
						},
						{
							Name:  "EMAIL_PASSWORD",
							Value: "",
						},
						{
							Name:  "EMAIL_USE_TLS",
							Value: "False",
						},
						{
							Name:  "EMAIL_USE_SSL",
							Value: "False",
						},
						{
							Name:  "DATABASE_HOST",
							Value: "postgres",
						},
						{
							Name:  "DATABASE_PORT",
							Value: "5432",
						},
						{
							Name:  "DATABASE_NAME",
							Value: "teamvault",
						},
						{
							Name:  "DATABASE_USER",
							Value: "teamvault",
						},
						{
							Name: "DATABASE_PASSWORD",
							ValueFrom: k8s.ValueFrom{
								SecretKeyRef: k8s.SecretKeyRef{
									Key:  "database-password",
									Name: "teamvault",
								},
							},
						},
						{
							Name: "SECRET_KEY",
							ValueFrom: k8s.ValueFrom{
								SecretKeyRef: k8s.SecretKeyRef{
									Key:  "secret-key",
									Name: "teamvault",
								},
							},
						},
						{
							Name: "FERNET_KEY",
							ValueFrom: k8s.ValueFrom{
								SecretKeyRef: k8s.SecretKeyRef{
									Key:  "fernet-key",
									Name: "teamvault",
								},
							},
						},
						{
							Name: "SALT",
							ValueFrom: k8s.ValueFrom{
								SecretKeyRef: k8s.SecretKeyRef{
									Key:  "salt",
									Name: "teamvault",
								},
							},
						},
						{
							Name:  "LDAP_ENABLED",
							Value: "true",
						},
						{
							Name:  "LDAP_SERVER_URI",
							Value: "ldap://ldap.ldap.svc.cluster.local",
						},
						{
							Name:  "LDAP_BIND_DN",
							Value: "cn=root,dc=benjamin-borbe,dc=de",
						},
						{
							Name: "LDAP_PASSWORD",
							ValueFrom: k8s.ValueFrom{
								SecretKeyRef: k8s.SecretKeyRef{
									Key:  "ldap-password",
									Name: "teamvault",
								},
							},
						},
						{
							Name:  "LDAP_USER_BASE_DN",
							Value: "ou=users,dc=benjamin-borbe,dc=de",
						},
						{
							Name:  "LDAP_USER_SEARCH_FILTER",
							Value: "(uid=%%(user)s)",
						},
						{
							Name:  "LDAP_GROUP_BASE_DN",
							Value: "ou=groups,dc=benjamin-borbe,dc=de",
						},
						{
							Name:  "LDAP_GROUP_SEARCH_FILTER",
							Value: "(objectClass=groupOfNames)",
						},
						{
							Name:  "LDAP_REQUIRE_GROUP",
							Value: "ou=employees,ou=groups,dc=benjamin-borbe,dc=de",
						},
						{
							Name:  "LDAP_ADMIN_GROUP",
							Value: "ou=admins,ou=groups,dc=benjamin-borbe,dc=de",
						},
						{
							Name:  "LDAP_ATTR_EMAIL",
							Value: "mail",
						},
						{
							Name:  "LDAP_ATTR_FIRST_NAME",
							Value: "givenName",
						},
						{
							Name:  "LDAP_ATTR_LAST_NAME",
							Value: "sn",
						},
					},
				},
				t.smtp().Container(),
			},
		},
		&deployer.ServiceDeployer{
			Context:   t.Cluster.Context,
			Namespace: "teamvault",
			Name:      "teamvault",
			Ports:     ports,
		},
		&deployer.IngressDeployer{
			Context:   t.Cluster.Context,
			Namespace: "teamvault",
			Name:      "teamvault",
			Port:      "http",
			Domains:   t.Domains,
		},
	}
}

func (t *Teamvault) smtp() *container.SmtpProvider {
	return &container.SmtpProvider{
		Hostname:     t.Domains[0].String(),
		Context:      t.Cluster.Context,
		Namespace:    "teamvault",
		SmtpPassword: t.SmtpPassword,
		SmtpUsername: t.SmtpUsername,
	}
}

func (t *Teamvault) Applier() (world.Applier, error) {
	return nil, nil
}
