package ssh

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/bborbe/run"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

type SSH struct {
	Host           Host
	PrivateKeyPath PrivateKeyPath
	User           User
}

type Host string

func (h Host) Validate(ctx context.Context) error {
	if h == "" {
		return errors.New("Host missing")
	}
	return nil
}

func (p Host) String() string {
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
	if s.Host == "" {
		return errors.New("Host missing")
	}
	if s.PrivateKeyPath == "" {
		return errors.New("PrivateKeyPath missing")
	}
	if s.User == "" {
		return errors.New("User missing")
	}
	return nil
}

func (s *SSH) Client(ctx context.Context) (*ssh.Client, error) {
	signer, err := s.PrivateKeyPath.Signer()
	if err != nil {
		return nil, errors.Wrap(err, "get signer failed")
	}
	client, err := ssh.Dial("tcp", s.Host.String(), &ssh.ClientConfig{
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
	client, err := s.Client(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get client failed")
	}
	session, err := client.NewSession()
	if err != nil {
		return nil, errors.Wrap(err, "create new ssh session failed")
	}
	return session, nil
}

func (s *SSH) Exists(ctx context.Context, path string) (bool, error) {
	session, err := s.createSession(ctx)
	if err != nil {
		return false, errors.Wrap(err, "create ssh session failed")
	}
	defer session.Close()
	return session.Run(fmt.Sprintf("stat %s", path)) == nil, nil
}

func (s *SSH) CreateFile(ctx context.Context, path string, content []byte) error {
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
				return err
			}
			stdin.Write(content)
			stdin.Close()
			return nil
		},
		func(ctx context.Context) error {
			return runWithout(session, fmt.Sprintf("cat > %s", path))
		},
	)
}

func runWithout(session *ssh.Session, cmd string) error {
	command := fmt.Sprintf("sudo sh -c 'export LANG=C; %s'", cmd)
	if err := session.Run(command); err != nil {
		return fmt.Errorf("run command '%s' failed", command)
	}
	return nil
}

func (s *SSH) ReadFile(ctx context.Context, path string) ([]byte, error) {
	session, err := s.createSession(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "create ssh session failed")
	}
	defer session.Close()
	return runOutput(ctx, session, fmt.Sprintf("cat %s", path))
}

func (s *SSH) Delete(ctx context.Context, path string) error {
	session, err := s.createSession(ctx)
	if err != nil {
		return errors.Wrap(err, "create ssh session failed")
	}
	defer session.Close()
	return runWithout(session, fmt.Sprintf("rm -rf %s", path))
}

func (s *SSH) CreateDir(ctx context.Context, path string) error {
	session, err := s.createSession(ctx)
	if err != nil {
		return errors.Wrap(err, "create ssh session failed")
	}
	defer session.Close()
	return runWithout(session, fmt.Sprintf("mkdir -p %s", path))
}

func (s *SSH) StartService(ctx context.Context, application string) error {
	session, err := s.createSession(ctx)
	if err != nil {
		return errors.Wrap(err, "create ssh session failed")
	}
	defer session.Close()
	return runWithout(session, fmt.Sprintf("systemctl start %s", application))
}

func (s *SSH) ServiceRunning(ctx context.Context, application string) (bool, error) {
	session, err := s.createSession(ctx)
	if err != nil {
		return false, errors.Wrap(err, "create ssh session failed")
	}
	defer session.Close()
	return session.Run(fmt.Sprintf("systemctl status %s", application)) == nil, nil
}
func runOutput(ctx context.Context, session *ssh.Session, cmd string) ([]byte, error) {
	b := &bytes.Buffer{}
	err := run.CancelOnFirstError(
		ctx,
		func(ctx context.Context) error {
			return runWithout(session, cmd)
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
