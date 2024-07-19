package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ruffel/godotreleaser/internal/cmd/root"
)

type exitCode int

const (
	exitOK     exitCode = 0
	exitError  exitCode = 1
	exitCancel exitCode = 2
)

func main() {
	os.Exit(int(mainRun()))
}

func mainRun() exitCode {
	if err := root.NewRootCmd().ExecuteContext(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)

		return exitError
	}

	return exitOK
}
