package kubectl

import (
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/runz0rd/jumper/cmd"
)

func Apply(resource string) (string, error) {
	fp := path.Join(os.TempDir(), "res.yaml")
	if err := os.WriteFile(fp, []byte(resource), 0644); err != nil {
		return "", err
	}
	return cmd.New("kubectl apply -f %v", fp).String()
}

func Create(resource string) (name string, err error) {
	fp := path.Join(os.TempDir(), "res.yaml")
	if err := os.WriteFile(fp, []byte(resource), 0644); err != nil {
		return "", err
	}
	out, err := cmd.New("kubectl create -f %v", fp).String()
	if err != nil {
		return "", errors.WithMessage(err, out)
	}
	return strings.TrimSuffix(out, " created\n"), nil
}

func Delete(resource string, wait bool) error {
	_, err := cmd.New("kubectl delete --wait=%v %v", wait, resource).String()
	return err
}

// resource needs to be in format resource/name eg pod/mypod01
func WaitResourceReady(namespace, resource, timeout string) error {
	if out, err := cmd.New("kubectl wait --for=condition=Ready %v --timeout=%v -n %v", resource, timeout, namespace).String(); err != nil {
		return errors.WithMessage(err, out)
	}
	return nil
}

func PortForward(namespace, resource string, srcPort, destPort int) *cmd.Cmd {
	return cmd.New("kubectl -n %v port-forward %v %v:%v", namespace, resource, srcPort, destPort)
}

func CopyToPod(namespace, pod, container, source, destination string) error {
	out, err := cmd.New("kubectl cp %v %v/%v:%v", source, namespace, pod, destination).String()
	if err != nil {
		return errors.WithMessage(err, out)
	}
	return nil
}
