package jumper

import (
	"bytes"
	"context"
	_ "embed"
	"os"
	"os/signal"
	"path"
	"strings"
	"text/template"

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

//go:embed template.config
var configTemplate string

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
	go handleExit(resource)
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
	idkeyServer, pubkeyServer, err := generateServerRSA(os.TempDir())
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
	log.Log().Infof("add public key to jumper server authorized keys")
	if err := AddPubkeyToPod(strings.TrimPrefix(resource, "pod/"), pubkeyServer); err != nil {
		return err
	}

	podIP, err := kubectl.GetResourceIP(resource)
	if err != nil {
		return err
	}
	config, err := renderConfig(configValues{
		IdentityFile: idkeyServer,
		HostName:     podIP,
	})
	if err != nil {
		return err
	}
	// # Using the SSH Server as a jumphost (via port-forward proxy), ssh into the desired Node
	log.Log().Infof("running ssh via proxy command")
	if err := sshViaProxy(config, sshArgs); err != nil {
		return err
	}
	return nil
}

func sshViaProxy(config string, sshArgs []string) error {
	c := []string{"ssh", "jumper", "-F", config}
	c = append(c, sshArgs...)
	out, err := cmd.New(c...).String()
	if err != nil {
		return errors.WithMessage(err, string(out))
	}
	return nil
}

func generateServerRSA(dir string) (idkey, pubkey string, err error) {
	idkey = path.Join(dir, "id_rsa")
	pubkey = path.Join(dir, "id_rsa.pub")
	os.Remove(idkey)
	os.Remove(pubkey)
	if out, err := cmd.New("ssh-keygen", "-t", "rsa", "-N", "", "-f", idkey).String(); err != nil {
		return idkey, pubkey, errors.WithMessage(err, out)
	}
	return idkey, pubkey, nil
}

func AddPubkeyToPod(pod, pubkeyServer string) error {
	if err := kubectl.CopyToPod(defaultNamespace, pod, "", pubkeyServer, "/id_rsa.pub"); err != nil {
		return err
	}
	return kubectl.ExecPod(pod, "cat /id_rsa.pub >> /root/.ssh/authorized_keys")
}

type configValues struct {
	IdentityFile string
	HostName     string
}

func renderTemplate(path string, values interface{}) (string, error) {
	tpl, err := template.ParseFiles(path)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	err = tpl.Execute(buf, values)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func renderConfig(c configValues) (string, error) {
	template, err := writeTempFile([]byte(configTemplate), "template.config")
	if err != nil {
		return "", err
	}
	configData, err := renderTemplate(template, c)
	if err != nil {
		return "", err
	}
	return writeTempFile([]byte(configData), "jumper.config")
}

func writeTempFile(data []byte, name string) (string, error) {
	fp := path.Join(os.TempDir(), name)
	err := os.WriteFile(fp, data, 0644)
	return fp, err
}

func handleExit(resource string) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	log.Log().Infof("exiting")
	kubectl.Delete(resource, false)
	os.Exit(0)
}
