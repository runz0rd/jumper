package jumper

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/bitfield/script"
	"github.com/pkg/errors"
	"github.com/runz0rd/jumper/kubectl"
)

const (
	defaultPod          = "jumper"
	defaultNamespace    = "default"
	defaultLocalSSHPort = 2222
)

//go:embed pod.yaml
var podYaml string

func Run(ctx context.Context, idkey, user, host string, port int, sshArgs []string) error {
	// create pod if not exists
	if _, err := kubectl.Apply(podYaml).
		WithStdout(os.Stdout).Stdout(); err != nil {
		return err
	}
	// wait for the pod to be ready
	if _, err := kubectl.WaitResourceReady(defaultNamespace, fmt.Sprintf("pod/%v", defaultPod), "90s").
		WithStdout(os.Stdout).Stdout(); err != nil {
		return err
	}
	// port forward into the pod
	go portForwardPod(ctx)

	// # Generate temp private/public key to ssh to the server
	idkeyServer, pubkeyServer, err := generateServerRSA(".")
	if err != nil {
		return err
	}
	// # Inject public SSH key to server
	if err := kubectl.CopyToPod(defaultNamespace, defaultPod, "", pubkeyServer, "/root/.ssh/authorized_keys"); err != nil {
		return err
	}
	// # Using the SSH Server as a jumphost (via port-forward proxy), ssh into the desired Node
	if err := sshViaProxy(idkey, user, host, idkeyServer, port, sshArgs); err != nil {
		return err
	}
	return nil
}

func sshViaProxy(idkey, user, host, idkeyServer string, port int, sshArgs []string) error {
	proxyCmd := fmt.Sprintf(`ssh root@127.0.0.1 -p %v -i %v %v \"%v\"`, defaultLocalSSHPort, idkeyServer, strings.Join(sshArgs, " "), "nc %h %p")
	cmd := fmt.Sprintf("ssh -i %v -p %v %v@%v -o ProxyCommand=%q %v", idkey, port, user, host, proxyCmd, strings.Join(sshArgs, " "))
	if out, err := script.Exec(cmd).String(); err != nil {
		return errors.WithMessage(err, out)
	}
	return nil
}

func generateServerRSA(dir string) (idkey, pubkey string, err error) {
	idkey = path.Join(dir, "id_rsa")
	pubkey = path.Join(dir, "id_rsa.pub")
	if out, err := script.Exec(fmt.Sprintf("ssh-keygen -t rsa -f %v -N ''", idkey)).String(); err != nil {
		return idkey, pubkey, errors.WithMessage(err, out)
	}
	return idkey, pubkey, nil
}

func portForwardPod(ctx context.Context) error {
	cmd := kubectl.PortForwardPod(ctx, defaultNamespace, defaultPod, defaultLocalSSHPort, 22)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_, err := cmd.Output()
	return err
}
