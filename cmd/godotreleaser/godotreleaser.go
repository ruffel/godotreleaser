package main

import (
	"context"
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/ruffel/godotreleaser/internal/cmd/root"
	"github.com/ruffel/godotreleaser/internal/godot/fetch"
)

type exitCode int

const (
	exitOK     exitCode = 0
	exitError  exitCode = 1
	exitCancel exitCode = 2
)

func main() {
	spew.Dump(fetch.BuildBinaryURL("3.2.3", false))

	os.Exit(int(mainRun()))
}

func mainRun() exitCode {
	if err := root.NewRootCmd().ExecuteContext(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)

		return exitError
	}

	return exitOK
}
