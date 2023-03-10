package main

import (
	"context"
	"log"
	"os"

	"github.com/runz0rd/jumper"
)

func main() {
	// debug := pflag.BoolP("debug", "d", false, "debug mode")
	// pflag.CommandLine.ParseErrorsWhitelist.UnknownFlags = true
	// pflag.Parse()

	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()
	defer ctx.Done()
	if err := jumper.Run(ctx, os.Args[1:], false); err != nil {
		log.Fatal(err)
	}
}
