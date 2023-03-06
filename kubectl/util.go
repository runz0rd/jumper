package kubectl

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/bitfield/script"
	"github.com/pkg/errors"
)

func Apply(resource string) *script.Pipe {
	return script.Exec(fmt.Sprintf("kubectl apply -f %v", resource))
}

// resource needs to be in format resource/name eg pod/mypod01
func WaitResourceReady(namespace, resource, timeout string) *script.Pipe {
	return script.Exec(fmt.Sprintf("kubectl wait --for=condition=Ready %v --timeout=%v -n %v", resource, timeout, namespace))
}

func PortForwardPod(ctx context.Context, namespace, pod string, srcPort, destPort int) *exec.Cmd {
	cmd := strings.Split(fmt.Sprintf("kubectl -n %v port-forward pods/%v %v:%v", namespace, pod, srcPort, destPort), " ")
	return exec.CommandContext(ctx, cmd[0], cmd[1:]...)
}

func CopyToPod(namespace, pod, container, source, destination string) error {
	out, err := script.Exec(fmt.Sprintf("kubectl -n %v cp %v %v:%v -c %v", namespace, source, pod, destination, container)).String()
	if err != nil {
		return errors.WithMessage(err, out)
	}
	return nil
}
