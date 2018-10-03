package ssh

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/bborbe/world/pkg/dns"

	"github.com/golang/glog"

	"github.com/bborbe/run"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

type SSH struct {
	Host           Host
	PrivateKeyPath PrivateKeyPath
	User           User
}

type Host struct {
	IP   dns.IP
	Port int
}

func (h Host) Address(ctx context.Context) (string, error) {
	ip, err := h.IP.IP(ctx)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%d", ip.String(), h.Port), nil
}

func (h Host) Validate(ctx context.Context) error {
	if err := h.IP.Validate(ctx); err != nil {
		return err
	}
	if h.Port <= 0 {
		return errors.New("invalid Port")
	}
	return nil
}

type PublicKeyPath string

func (h PublicKeyPath) Validate(ctx context.Context) error {
	if h == "" {
		return errors.New("PublicKeyPath missing")
	}
	return nil
}

func (p PublicKeyPath) String() string {
	return string(p)
}

type PrivateKeyPath string

func (h PrivateKeyPath) Validate(ctx context.Context) error {
	if h == "" {
		return errors.New("PrivateKeyPath missing")
	}
	return nil
}

func (p PrivateKeyPath) String() string {
	return string(p)
}

func (p PrivateKeyPath) Signer() (ssh.Signer, error) {
	content, err := ioutil.ReadFile(p.String())
	if err != nil {
		return nil, errors.Wrap(err, "read private key failed")
	}
	signer, err := ssh.ParsePrivateKey(content)
	if err != nil {
		return nil, errors.Wrap(err, "parse private key failed")
	}
	return signer, nil
}

type User string

func (h User) Validate(ctx context.Context) error {
	if h == "" {
		return errors.New("User missing")
	}
	return nil
}

func (p User) String() string {
	return string(p)
}

func (s SSH) Validate(ctx context.Context) error {
	if err := s.Host.Validate(ctx); err != nil {
		return err
	}
	if s.PrivateKeyPath == "" {
		return errors.New("PrivateKeyPath missing")
	}
	if s.User == "" {
		return errors.New("User missing")
	}
	return nil
}

func (s *SSH) client(ctx context.Context) (*ssh.Client, error) {
	signer, err := s.PrivateKeyPath.Signer()
	if err != nil {
		return nil, errors.Wrap(err, "get signer failed")
	}
	addr, err := s.Host.Address(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get addr from host failed")
	}
	client, err := ssh.Dial("tcp", addr, &ssh.ClientConfig{
		User: s.User.String(),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		return nil, errors.Wrap(err, "connect host failed")
	}
	return client, nil
}

func (s *SSH) createSession(ctx context.Context) (*ssh.Session, error) {
	client, err := s.client(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get client failed")
	}
	session, err := client.NewSession()
	if err != nil {
		return nil, errors.Wrap(err, "create new ssh session failed")
	}
	return session, nil
}

func (s *SSH) RunCommandStdin(ctx context.Context, command string, content []byte) error {
	session, err := s.createSession(ctx)
	if err != nil {
		return errors.Wrap(err, "create ssh session failed")
	}
	defer session.Close()

	return run.CancelOnFirstError(
		ctx,
		func(ctx context.Context) error {
			stdin, err := session.StdinPipe()
			if err != nil {
				return errors.Wrap(err, "open stdinpipe failed")
			}
			defer stdin.Close()
			stdin.Write(content)
			glog.V(2).Infof("write content to stdin complete")
			return nil
		},
		func(ctx context.Context) error {
			return runWithout(ctx, session, command)
		},
	)
}

func (s *SSH) RunCommandStdout(ctx context.Context, command string) ([]byte, error) {
	session, err := s.createSession(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "create ssh session failed")
	}
	defer session.Close()

	b := &bytes.Buffer{}
	err = run.CancelOnFirstError(
		ctx,
		func(ctx context.Context) error {
			return runWithout(ctx, session, command)
		},
		func(ctx context.Context) error {
			stdin, err := session.StdoutPipe()
			if err != nil {
				return err
			}
			io.Copy(b, stdin)
			return nil
		},
	)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (s *SSH) RunCommand(ctx context.Context, cmd string) error {
	session, err := s.createSession(ctx)
	if err != nil {
		return errors.Wrap(err, "create ssh session failed")
	}
	defer session.Close()
	return runWithout(ctx, session, cmd)
}

func runWithout(ctx context.Context, session *ssh.Session, cmd string) error {
	command := fmt.Sprintf("sudo sh -c 'export LANG=C; %s'", cmd)
	glog.V(1).Infof("run remote command: %s", command)
	select {
	case <-ctx.Done():
		glog.V(1).Infof("context done => send kill")
		return session.Signal(ssh.SIGKILL)
	default:
		return errors.Wrapf(session.Run(command), "run command failed: %s", command)
	}
}
