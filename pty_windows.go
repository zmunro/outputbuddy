//go:build windows
// +build windows

package main

func runWithPty(router *OutputRouter, commandArgs []string) error {
	// PTY is not well supported on Windows, fall back to pipe mode
	return runWithPipes(router, commandArgs)
}
