//go:build !windows
// +build !windows

package main

import (
	"bufio"
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

	scanner := bufio.NewScanner(ptmx)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	scanner.Split(scanLinesOrProgress)

	for scanner.Scan() {
		data := scanner.Bytes()
		router.WriteStdout(append(data, '\n'))
	}

	return cmd.Wait()
}
