// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"math/big"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/bborbe/world/pkg/apt"
	"github.com/bborbe/world/pkg/remote"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type OpenvpnIP string

func (o OpenvpnIP) String() string {
	return string(o)
}

func (o OpenvpnIP) Validate(ctx context.Context) error {
	if o == "" {
		return errors.New("OpenvpnIP empty")
	}
	return nil
}

type OpenvpnHostname string

func (o OpenvpnHostname) String() string {
	return string(o)
}
func (o OpenvpnHostname) Validate(ctx context.Context) error {
	if o == "" {
		return errors.New("OpenvpnHostname empty")
	}
	return nil
}

type OpenvpnRoutes []OpenvpnRoute

func (o OpenvpnRoutes) Validate(ctx context.Context) error {
	for _, route := range o {
		if err := route.Validate(ctx); err != nil {
			return err
		}
	}
	return nil
}

type OpenvpnRoute struct {
	Gatway OpenvpnIP
	IP     OpenvpnIP
	Mask   OpenvpnIP
}

func (o OpenvpnRoute) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		o.Mask,
		o.IP,
		o.Gatway,
	)
}

type OpenvpnServer struct {
	SSH        *ssh.SSH
	ServerName OpenvpnHostname
	ServerIP   OpenvpnIP
	ServerMask OpenvpnIP
	Routes     OpenvpnRoutes
}

func (o *OpenvpnServer) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		o.SSH,
		o.ServerName,
		o.ServerIP,
		o.ServerMask,
		o.Routes,
	)
}

func (o *OpenvpnServer) Children() []world.Configuration {
	return []world.Configuration{
		&File{
			SSH:     o.SSH,
			Path:    "/etc/default/openvpn",
			User:    "root",
			Group:   "root",
			Perm:    0644,
			Content: remote.StaticContent(openvpnDefaultConf),
		},
		&Directory{
			SSH:   o.SSH,
			Path:  "/etc/openvpn/keys",
			User:  "root",
			Group: "root",
			Perm:  0700,
		},
		&File{
			SSH:     o.SSH,
			Path:    "/etc/openvpn/server.conf",
			User:    "root",
			Group:   "root",
			Perm:    0600,
			Content: remote.ContentFunc(o.serverConfigContent),
		},
		&File{
			SSH:     o.SSH,
			Path:    "/etc/openvpn/keys/ta.key",
			User:    "root",
			Group:   "root",
			Perm:    0600,
			Content: remote.ContentFunc(o.taKey),
		},
		&File{
			SSH:     o.SSH,
			Path:    "/etc/openvpn/keys/dh.pem",
			User:    "root",
			Group:   "root",
			Perm:    0600,
			Content: remote.ContentFunc(o.dhPem),
		},
		&File{
			SSH:     o.SSH,
			Path:    "/etc/openvpn/keys/ca.crt",
			User:    "root",
			Group:   "root",
			Perm:    0600,
			Content: remote.ContentFunc(o.caCrt),
		},
		&File{
			SSH:     o.SSH,
			Path:    "/etc/openvpn/keys/server.crt",
			User:    "root",
			Group:   "root",
			Perm:    0600,
			Content: remote.ContentFunc(o.serverCrt),
		},
		&File{
			SSH:     o.SSH,
			Path:    "/etc/openvpn/keys/server.key",
			User:    "root",
			Group:   "root",
			Perm:    0600,
			Content: remote.ContentFunc(o.serverKey),
		},
		world.NewConfiguraionBuilder().WithApplier(&remote.Iptables{
			SSH:  o.SSH,
			Port: 563,
		}),
		world.NewConfiguraionBuilder().WithApplier(&apt.Update{
			SSH: o.SSH,
		}),
		world.NewConfiguraionBuilder().WithApplier(&apt.Install{
			SSH:     o.SSH,
			Package: "openvpn",
		}),
		world.NewConfiguraionBuilder().WithApplier(&apt.Autoremove{
			SSH: o.SSH,
		}),
		world.NewConfiguraionBuilder().WithApplier(&apt.Clean{
			SSH: o.SSH,
		}),
		world.NewConfiguraionBuilder().WithApplier(&remote.ServiceStart{
			SSH:  o.SSH,
			Name: "openvpn",
		}),
	}
}

func (o *OpenvpnServer) Applier() (world.Applier, error) {
	return nil, nil
}

func (o *OpenvpnServer) serverConfigContent(ctx context.Context) ([]byte, error) {
	type Route struct {
		Gateway string
		IP      string
		Netmask string
	}
	data := struct {
		ServerIP      string
		ServerNetmask string
		Routes        []Route
	}{
		ServerIP:      o.ServerIP.String(),
		ServerNetmask: o.ServerMask.String(),
		Routes:        []Route{},
	}
	for _, route := range o.Routes {
		data.Routes = append(data.Routes, Route{
			Gateway: route.Gatway.String(),
			IP:      route.IP.String(),
			Netmask: route.Mask.String(),
		})
	}

	return render(openvpnServerConfig, data)
}

const openvpnServerConfig = `
dev tap0

proto tcp-server
port 563

server {{.ServerIP}} {{.ServerNetmask}}
ifconfig-pool-persist ip_pool

mode server
status server.status
tls-auth /etc/openvpn/keys/ta.key 0
keepalive 10 30
client-to-client
max-clients 150
verb 3

tls-server
dh /etc/openvpn/keys/dh.pem
ca /etc/openvpn/keys/ca.crt
cert /etc/openvpn/keys/server.crt
key /etc/openvpn/keys/server.key
comp-lzo

persist-key
persist-tun

log /var/log/openvpn/server.log

{{range $route := .Routes}}
route {{$route.IP}} {{$route.Netmask}} {{$route.Gateway}}
{{ end }} 
`

func (o *OpenvpnServer) certDirectory(ctx context.Context) (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", errors.Wrap(err, "get homedir failed")
	}
	return filepath.Join(usr.HomeDir, ".openvpn", o.ServerName.String()), nil
}

func (o *OpenvpnServer) readOrCreate(
	ctx context.Context,
	filename string,
	generateFunc func(ctx context.Context) ([]byte, error),
) ([]byte, error) {
	certDirectory, err := o.certDirectory(ctx)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(certDirectory, 0700); err != nil {
		return nil, err
	}
	path := filepath.Join(certDirectory, filename)
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		glog.V(2).Infof("%s not existing => generate", path)
		content, err := generateFunc(ctx)
		if err != nil {
			return nil, err
		}
		ioutil.WriteFile(path, content, 0600)
		return content, nil
	}
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func (o *OpenvpnServer) caPrivateKey(ctx context.Context) ([]byte, error) {
	return o.readOrCreate(ctx, "ca.key", func(ctx context.Context) ([]byte, error) {
		caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			return nil, err
		}
		return pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
		}), nil
	})
}

func (o *OpenvpnServer) caCertifcate() *x509.Certificate {
	return &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			Organization:  []string{"Benjamin Borbe"},
			Country:       []string{"DE"},
			Province:      []string{"Hessen"},
			Locality:      []string{"Wiesbaden"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
}
func (o *OpenvpnServer) caCrt(ctx context.Context) ([]byte, error) {
	return o.readOrCreate(ctx, "ca.crt", func(ctx context.Context) ([]byte, error) {
		caPriv, err := o.caPrivateKey(ctx)
		if err != nil {
			return nil, err
		}

		caPrivPem, _ := pem.Decode(caPriv)
		if caPrivPem.Type != "RSA PRIVATE KEY" {
			return nil, errors.Errorf("invalid type %s", caPrivPem.Type)
		}

		caPrivKey, err := x509.ParsePKCS1PrivateKey(caPrivPem.Bytes)
		if err != nil {
			return nil, err
		}

		caCert, err := x509.CreateCertificate(rand.Reader, o.caCertifcate(), o.caCertifcate(), &caPrivKey.PublicKey, caPrivKey)
		if err != nil {
			return nil, err
		}

		return pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: caCert,
		}), nil
	})
}

func (o *OpenvpnServer) serverCrt(ctx context.Context) ([]byte, error) {
	return o.readOrCreate(ctx, "server.crt", func(ctx context.Context) ([]byte, error) {
		cert := &x509.Certificate{
			SerialNumber: big.NewInt(1658),
			Subject: pkix.Name{
				Organization:  []string{"Benjamin Borbe"},
				Country:       []string{"DE"},
				Province:      []string{"Hessen"},
				Locality:      []string{"Wiesbaden"},
				StreetAddress: []string{""},
				PostalCode:    []string{""},
			},
			NotBefore:    time.Now(),
			NotAfter:     time.Now().AddDate(10, 0, 0),
			SubjectKeyId: []byte{1, 2, 3, 4, 6},
			ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
			KeyUsage:     x509.KeyUsageDigitalSignature,
		}

		caPriv, err := o.caPrivateKey(ctx)
		if err != nil {
			return nil, err
		}

		caPrivPem, _ := pem.Decode(caPriv)
		if caPrivPem.Type != "RSA PRIVATE KEY" {
			return nil, errors.Errorf("invalid type %s", caPrivPem.Type)
		}

		caPrivKey, err := x509.ParsePKCS1PrivateKey(caPrivPem.Bytes)
		if err != nil {
			return nil, err
		}

		certKey, err := o.serverKey(ctx)
		if err != nil {
			return nil, err
		}

		certPrivPem, _ := pem.Decode(certKey)
		if certPrivPem.Type != "RSA PRIVATE KEY" {
			return nil, errors.Errorf("invalid type %s", certPrivPem.Type)
		}

		certPrivKey, err := x509.ParsePKCS1PrivateKey(certPrivPem.Bytes)
		if err != nil {
			return nil, err
		}

		certBytes, err := x509.CreateCertificate(rand.Reader, cert, o.caCertifcate(), &certPrivKey.PublicKey, caPrivKey)
		if err != nil {
			return nil, err
		}

		return pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: certBytes,
		}), nil
	})
}

func (o *OpenvpnServer) serverKey(ctx context.Context) ([]byte, error) {
	return o.readOrCreate(ctx, "server.key", func(ctx context.Context) ([]byte, error) {
		caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			return nil, err
		}
		return pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
		}), nil
	})
}

func (o *OpenvpnServer) taKey(ctx context.Context) ([]byte, error) {
	return o.readOrCreate(ctx, "ta.key", func(ctx context.Context) ([]byte, error) {
		caPrivKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, err
		}
		return pem.EncodeToMemory(&pem.Block{
			Type:  "OpenVPN Static key V1",
			Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
		}), nil
	})
}

func (o *OpenvpnServer) dhPem(ctx context.Context) ([]byte, error) {
	return o.readOrCreate(ctx, "dh.pem", func(ctx context.Context) ([]byte, error) {
		priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, err
		}
		content, err := x509.MarshalECPrivateKey(priv)
		if err != nil {
			return nil, err
		}
		return pem.EncodeToMemory(&pem.Block{
			Type:  "DH PARAMETERS",
			Bytes: content,
		}), nil
	})
}

var openvpnDefaultConf = []byte(`
AUTOSTART="all"
OPTARGS=""
OMIT_SENDSIGS=0
`)
