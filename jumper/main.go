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

	debug := false
	debugEnv := os.Getenv("DEBUG")
	if debugEnv == "1" || debugEnv == "true" {
		debug = true
	}
	if err := jumper.Run(ctx, os.Args[1:], debug); err != nil {
		log.Fatal(err)
	}
}
