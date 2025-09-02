//go:build windows
// +build windows

package main

import (
	"os"
)

func handleWindowResize(ptmx *os.File) {
	// Window resizing is not supported on Windows
	// PTY support on Windows is limited
}
