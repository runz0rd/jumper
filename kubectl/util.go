package kubectl

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/runz0rd/jumper/cmd"
)

func Apply(resource string) (string, error) {
	fp := filepath.Join(os.TempDir(), "res.yaml")
	if err := os.WriteFile(fp, []byte(resource), 0644); err != nil {
		return "", err
	}
	return cmd.New("kubectl", "apply", "-f", fp).String()
}

func Create(resource string) (name string, err error) {
	fp := filepath.Join(os.TempDir(), "res.yaml")
	if err := os.WriteFile(fp, []byte(resource), 0644); err != nil {
		return "", err
	}
	out, err := cmd.New("kubectl", "create", "-f", fp).String()
	if err != nil {
		return "", errors.WithMessage(err, out)
	}
	return strings.TrimSuffix(out, " created\n"), nil
}

func Delete(resource string, wait bool) error {
	_, err := cmd.New("kubectl", "delete", fmt.Sprintf("--wait=%v", wait), resource).String()
	return err
}

// resource needs to be in format resource/name eg pod/mypod01
func WaitResourceReady(namespace, resource, timeout string) error {
	if out, err := cmd.New("kubectl", "wait", "--for=condition=Ready", resource, fmt.Sprintf("--timeout=%v", timeout), "-n", namespace).String(); err != nil {
		return errors.WithMessage(err, out)
	}
	return nil
}

func PortForward(namespace, resource string, srcPort, destPort int) *cmd.Cmd {
	return cmd.New("kubectl", "-n", namespace, "port-forward", resource, fmt.Sprintf("%v:%v", srcPort, destPort))
}

func ExecPod(pod, c string) error {
	if out, err := cmd.New("kubectl", "exec", pod, "--", "bash", "-c", c).String(); err != nil {
		return errors.WithMessage(err, out)
	}
	return nil
}

func CopyToPod(namespace, pod, container, source, destination string) error {
	out, err := cmd.New("kubectl", "cp", source, fmt.Sprintf("%v/%v:%v", namespace, pod, destination)).String()
	if err != nil {
		return errors.WithMessage(err, out)
	}
	return nil
}

func GetResourceIP(resource string) (string, error) {
	// kubectl get pod nginx --template '{{.status.podIP}}'
	return cmd.New("kubectl", "get", resource, "--template", "{{.status.podIP}}").String()
}
