package main

import (
	"context"
	"log"

	"github.com/runz0rd/jumper"
	"github.com/spf13/pflag"
)

func main() {
	var idkey, user, host string
	var port int
	var sshArgs []string
	pflag.StringVarP(&idkey, "identity", "i", "", "identity key file, or PEM")
	pflag.StringVarP(&user, "user", "u", "", "ssh username")
	pflag.IntVarP(&port, "port", "p", 22, "ssh port for target server")
	pflag.Parse()

	host = pflag.Arg(0)
	sshArgs = pflag.Args()[1:]
	// spew.Dump(idkey, user, host, port, sshArgs)
	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()
	if err := jumper.Run(ctx, idkey, user, host, port, sshArgs); err != nil {
		log.Fatal(err)
	}
}
