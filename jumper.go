package jumper

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/runz0rd/jumper/cmd"
	"github.com/runz0rd/jumper/kubectl"
	"github.com/runz0rd/jumper/log"
	"github.com/sirupsen/logrus"
)

const (
	defaultPod          = "jumper"
	defaultNamespace    = "default"
	defaultLocalSSHPort = 2222
)

//go:embed pod.yaml
var podYaml string

func Run(ctx context.Context, sshArgs []string, debug bool) error {
	if debug {
		log.SetLevel(logrus.DebugLevel)
	}
	// create pod if not exists
	log.Log().Infof("creating jumper pod if it doesnt exist")
	resource, err := kubectl.Create(podYaml)
	if err != nil {
		return err
	}
	defer kubectl.Delete(resource, false)
	// wait for the pod to be ready
	log.Log().Infof("waiting for %v to get ready", resource)
	if err := kubectl.WaitResourceReady(defaultNamespace, resource, "90s"); err != nil {
		return err
	}
	// port forward into the pod
	log.Log().Infof("port-forwarding %v to %v", defaultLocalSSHPort, 22)

	portForwardCmd := kubectl.PortForward(defaultNamespace, resource, defaultLocalSSHPort, 22)
	go portForwardCmd.String()
	defer portForwardCmd.Kill()

	// # Generate temp private/public key to ssh to the server
	log.Log().Infof("generating private/public keys for jumper server")
	idkeyServer, pubkeyServer, err := generateServerRSA(".")
	if err != nil {
		return err
	}
	defer func() {
		if err := os.Remove(idkeyServer); err != nil {
			log.Log().Infof("error removing server keys: %v", idkeyServer)
		}
		if err := os.Remove(pubkeyServer); err != nil {
			log.Log().Infof("error removing server keys: %v", pubkeyServer)
		}
	}()
	// # Inject public SSH key to server
	log.Log().Infof("copyng public key to jumper server authorized keys")
	if err := kubectl.CopyToPod(defaultNamespace, strings.TrimPrefix(resource, "pod/"), "", pubkeyServer, "/root/.ssh/authorized_keys"); err != nil {
		return err
	}
	return nil
	// # Using the SSH Server as a jumphost (via port-forward proxy), ssh into the desired Node
	log.Log().Infof("running ssh via proxy command")
	if err := sshViaProxy(idkeyServer, sshArgs); err != nil {
		return err
	}
	return nil
}

func sshViaProxy(idkeyServer string, sshArgs []string) error {
	proxyCmd := fmt.Sprintf("ssh root@127.0.0.1 -p %v -i %v  %q", defaultLocalSSHPort, idkeyServer, "nc %h %p")
	_, err := cmd.New("ssh %v ProxyCommand='%v'", strings.Join(sshArgs, " "), proxyCmd).String()
	if err != nil {
		return err
	}
	return nil
}

func generateServerRSA(dir string) (idkey, pubkey string, err error) {
	idkey = path.Join(dir, "id_rsa")
	pubkey = path.Join(dir, "id_rsa.pub")
	_, errIdkey := os.Stat(idkey)
	_, errPubkey := os.Stat(pubkey)
	if errIdkey == nil && errPubkey == nil {
		return idkey, pubkey, nil
	}
	if out, err := cmd.New(`ssh-keygen -t rsa -f %v -N ''`, idkey).String(); err != nil {
		return idkey, pubkey, errors.WithMessage(err, out)
	}
	return idkey, pubkey, nil
}
