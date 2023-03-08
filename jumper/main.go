package main

import (
	"context"
	"log"

	"github.com/runz0rd/jumper"
	"github.com/spf13/pflag"
)

func main() {
	debug := pflag.BoolP("debug", "d", false, "debug mode")
	pflag.CommandLine.ParseErrorsWhitelist.UnknownFlags = true
	pflag.Parse()

	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()
	defer ctx.Done()
	if err := jumper.Run(ctx, pflag.Args(), *debug); err != nil {
		log.Fatal(err)
	}
}
