//go:build !windows
// +build !windows

package main

import (
	"io"
	"os"
	"os/exec"

	"github.com/creack/pty"
	"golang.org/x/term"
)

func runWithPty(router *OutputRouter, commandArgs []string) error {
	cmd := exec.Command(commandArgs[0], commandArgs[1:]...)

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return err
	}
	defer ptmx.Close()

	handleWindowResize(ptmx)

	if term.IsTerminal(int(os.Stdin.Fd())) {
		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err == nil {
			defer term.Restore(int(os.Stdin.Fd()), oldState)
		}
	}

	go io.Copy(ptmx, os.Stdin)

	// Read PTY output in chunks and pass directly to router
	// This preserves carriage returns and allows progress bars to work correctly
	buf := make([]byte, 32*1024)
	for {
		n, err := ptmx.Read(buf)
		if n > 0 {
			router.WriteStdout(buf[:n])
		}
		if err != nil {
			break
		}
	}

	return cmd.Wait()
}
