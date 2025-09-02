//go:build !windows
// +build !windows

package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/creack/pty"
)

func handleWindowResize(ptmx *os.File) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			pty.InheritSize(os.Stdin, ptmx)
		}
	}()
	ch <- syscall.SIGWINCH
}
