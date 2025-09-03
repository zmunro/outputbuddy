package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"golang.org/x/term"
)

var VERSION = "2.1.0"

// FileWriter handles writing to a file with optional ANSI stripping
type FileWriter struct {
	file      *os.File
	stripAnsi bool
	buffer    bytes.Buffer
	lastLine  string
	currentLine bytes.Buffer  // Accumulates the current line being built
}

func NewFileWriter(path string, stripAnsi bool) (*FileWriter, error) {
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	return &FileWriter{
		file:      file,
		stripAnsi: stripAnsi,
	}, nil
}

func (fw *FileWriter) Write(p []byte) (n int, err error) {
	// Debug: log data received
	// fmt.Fprintf(os.Stderr, "[FW] Write %d bytes: %q\n", len(p), string(p))

	if !fw.stripAnsi {
		// For raw output, we still need to handle carriage returns
		// to avoid duplicate progress bar lines
		fw.buffer.Write(p)
		return len(p), fw.processBuffer(false)
	}

	fw.buffer.Write(p)
	return len(p), fw.processBuffer(true)
}

func (fw *FileWriter) processBuffer(stripAnsi bool) error {
	for fw.buffer.Len() > 0 {
		b := fw.buffer.Bytes()

		// Find the next line ending or carriage return
		newlineIdx := bytes.IndexByte(b, '\n')
		crIdx := bytes.IndexByte(b, '\r')

		if newlineIdx == -1 && crIdx == -1 {
			// No line ending found, accumulate in currentLine
			fw.currentLine.Write(b)
			fw.buffer.Reset()
			break
		}

		// Determine which comes first
		idx := newlineIdx
		isCR := false
		if crIdx != -1 && (newlineIdx == -1 || crIdx < newlineIdx) {
			idx = crIdx
			isCR = true
		}

		// Add everything up to the line ending to currentLine
		fw.currentLine.Write(b[:idx])

		if isCR {
			// Carriage return: check if it's followed by newline (CRLF)
			fw.buffer.Next(idx + 1)
			remaining := fw.buffer.Bytes()
			if len(remaining) > 0 && remaining[0] == '\n' {
				// CRLF sequence - treat as newline
				fw.buffer.Next(1) // consume the \n
				line := fw.currentLine.Bytes()
				fw.currentLine.Reset()

				if stripAnsi {
					line = fw.cleanLine(line)
					trimmed := bytes.TrimSpace(line)
					if len(trimmed) == 0 {
						continue
					}

					if fw.isProgressLine(line) {
						fw.lastLine = string(line)
						continue
					}
				}

				fw.file.Write(line)
				fw.file.Write([]byte("\n"))
				fw.lastLine = string(line)
			} else {
				// Standalone CR - reset current line (for progress bars)
				fw.currentLine.Reset()
			}
			continue
		}

		// Newline: write the accumulated line to file
		line := fw.currentLine.Bytes()
		fw.currentLine.Reset()
		fw.buffer.Next(idx + 1)

		if stripAnsi {
			line = fw.cleanLine(line)
			trimmed := bytes.TrimSpace(line)
			if len(trimmed) == 0 {
				continue
			}

			if fw.isProgressLine(line) {
				fw.lastLine = string(line)
				continue
			}
		}

		fw.file.Write(line)
		fw.file.Write([]byte("\n"))
		fw.lastLine = string(line)
	}

	return nil
}

func (fw *FileWriter) cleanLine(line []byte) []byte {
	cleaned := removeAnsiSequences(line)

	cleaned = bytes.ReplaceAll(cleaned, []byte("\r"), []byte{})

	cleaned = removeBrailleChars(cleaned)

	trimmed := bytes.TrimSpace(cleaned)
	if len(trimmed) == 0 {
		return []byte{}
	}

	return cleaned
}

func (fw *FileWriter) isProgressLine(line []byte) bool {
	lineStr := string(line)

	if strings.Contains(lineStr, "s Run") || strings.Contains(lineStr, "s Build") {
		return true
	}

	trimmed := strings.TrimSpace(lineStr)
	if strings.HasPrefix(trimmed, "0.") || strings.HasPrefix(trimmed, "1.") ||
	   strings.HasPrefix(trimmed, "2.") || strings.HasPrefix(trimmed, "3.") {
		return true
	}

	return false
}

func (fw *FileWriter) Flush() error {
	// fmt.Fprintf(os.Stderr, "[FW] Flush called, buffer=%d bytes, currentLine=%d bytes\n", fw.buffer.Len(), fw.currentLine.Len())
	// Process any remaining buffered data
	fw.processBuffer(fw.stripAnsi)

	// Write any incomplete line
	if fw.currentLine.Len() > 0 {
		line := fw.currentLine.Bytes()
		if fw.stripAnsi {
			line = fw.cleanLine(line)
			// After cleaning, check if there's still content
			if len(bytes.TrimSpace(line)) > 0 {
				fw.file.Write(line)
				fw.file.Write([]byte("\n"))
			}
		} else {
			// For non-stripped output, write as-is
			fw.file.Write(line)
			fw.file.Write([]byte("\n"))
		}
		fw.currentLine.Reset()
	}

	return fw.file.Sync()
}

func (fw *FileWriter) Close() error {
	fw.Flush()
	return fw.file.Close()
}

func removeAnsiSequences(data []byte) []byte {
	result := data

	// CSI sequences: ESC [ ... letter
	for {
		start := bytes.Index(result, []byte("\x1b["))
		if start == -1 {
			break
		}

		end := start + 2
		for end < len(result) {
			b := result[end]
			if (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') {
				end++
				break
			}
			end++
		}

		result = append(result[:start], result[end:]...)
	}

	// OSC sequences: ESC ] ... BEL or ESC \
	for {
		start := bytes.Index(result, []byte("\x1b]"))
		if start == -1 {
			break
		}

		end := bytes.Index(result[start:], []byte("\x07"))
		if end == -1 {
			end = bytes.Index(result[start:], []byte("\x1b\\"))
			if end == -1 {
				break
			}
			end += 2
		} else {
			end++
		}

		result = append(result[:start], result[start+end:]...)
	}

	for {
		idx := bytes.IndexByte(result, '\x1b')
		if idx == -1 || idx >= len(result)-1 {
			break
		}
		result = append(result[:idx], result[idx+2:]...)
	}

	return result
}

func removeBrailleChars(data []byte) []byte {
	var result []byte
	for len(data) > 0 {
		r, size := utf8DecodeRune(data)
		if r < 0x2800 || r > 0x28FF {
			result = append(result, data[:size]...)
		}
		data = data[size:]
	}
	return result
}

// Simple UTF-8 decoder for Braille detection
func utf8DecodeRune(p []byte) (rune, int) {
	if len(p) == 0 {
		return 0, 0
	}

	b0 := p[0]
	if b0 < 0x80 {
		return rune(b0), 1
	}

	if len(p) < 2 || b0 < 0xC0 {
		return rune(b0), 1
	}

	if b0 < 0xE0 {
		if len(p) < 2 {
			return rune(b0), 1
		}
		return rune(b0&0x1F)<<6 | rune(p[1]&0x3F), 2
	}

	if b0 < 0xF0 {
		if len(p) < 3 {
			return rune(b0), 1
		}
		return rune(b0&0x0F)<<12 | rune(p[1]&0x3F)<<6 | rune(p[2]&0x3F), 3
	}

	if len(p) < 4 {
		return rune(b0), 1
	}
	return rune(b0&0x07)<<18 | rune(p[1]&0x3F)<<12 | rune(p[2]&0x3F)<<6 | rune(p[3]&0x3F), 4
}

// TerminalWriter wraps terminal output to handle progress bars correctly
type TerminalWriter struct {
	output io.Writer
	isStderr bool
}

func NewTerminalWriter(output io.Writer, isStderr bool) *TerminalWriter {
	return &TerminalWriter{
		output: output,
		isStderr: isStderr,
	}
}

func (tw *TerminalWriter) Write(p []byte) (n int, err error) {
	// Write directly to terminal, preserving carriage returns for proper progress bar behavior
	return tw.output.Write(p)
}

// OutputRouter manages routing output to multiple destinations
type OutputRouter struct {
	stdoutWriters []io.Writer
	stderrWriters []io.Writer
	fileWriters   []*FileWriter
	mu            sync.Mutex
}

func NewOutputRouter() *OutputRouter {
	return &OutputRouter{}
}

func (or *OutputRouter) AddStdoutFile(path string, stripAnsi bool) error {
	fw, err := NewFileWriter(path, stripAnsi)
	if err != nil {
		return err
	}
	or.stdoutWriters = append(or.stdoutWriters, fw)
	or.fileWriters = append(or.fileWriters, fw)
	return nil
}

func (or *OutputRouter) AddStderrFile(path string, stripAnsi bool) error {
	fw, err := NewFileWriter(path, stripAnsi)
	if err != nil {
		return err
	}
	or.stderrWriters = append(or.stderrWriters, fw)
	or.fileWriters = append(or.fileWriters, fw)
	return nil
}

func (or *OutputRouter) AddCombinedFile(path string, stripAnsi bool) error {
	fw, err := NewFileWriter(path, stripAnsi)
	if err != nil {
		return err
	}
	or.stdoutWriters = append(or.stdoutWriters, fw)
	or.stderrWriters = append(or.stderrWriters, fw)
	or.fileWriters = append(or.fileWriters, fw)
	return nil
}

func (or *OutputRouter) AddStdoutTerminal() {
	or.stdoutWriters = append(or.stdoutWriters, NewTerminalWriter(os.Stdout, false))
}

func (or *OutputRouter) AddStderrTerminal() {
	or.stderrWriters = append(or.stderrWriters, NewTerminalWriter(os.Stderr, true))
}

func (or *OutputRouter) WriteStdout(data []byte) {
	or.mu.Lock()
	defer or.mu.Unlock()

	for _, w := range or.stdoutWriters {
		w.Write(data)
	}
}

func (or *OutputRouter) WriteStderr(data []byte) {
	or.mu.Lock()
	defer or.mu.Unlock()

	for _, w := range or.stderrWriters {
		w.Write(data)
	}
}

func (or *OutputRouter) Close() {
	for _, fw := range or.fileWriters {
		fw.Close()
	}
}

// Parse arguments and setup router
func parseArgs(args []string) (*OutputRouter, []string, bool, error) {
	router := NewOutputRouter()
	stripAnsi := true
	usePty := true

	dashIndex := -1
	for i, arg := range args {
		if arg == "--" {
			dashIndex = i
			break
		}
	}

	if dashIndex == -1 {
		return nil, nil, false, fmt.Errorf("no -- separator found")
	}

	routerArgs := args[:dashIndex]
	commandArgs := args[dashIndex+1:]

	if len(commandArgs) == 0 {
		return nil, nil, false, fmt.Errorf("no command specified after --")
	}

		fileMap := make(map[string]bool)

	// First pass: check for flags
	filteredArgs := []string{}
	for _, arg := range routerArgs {
		if arg == "--no-pty" {
			usePty = false
			continue
		}
		if arg == "--keep-ansi" {
			stripAnsi = false
			continue
		}
		filteredArgs = append(filteredArgs, arg)
	}

	// Default behavior: if no routing args provided (after filtering flags), redirect both to buddy.log and terminal
	if len(filteredArgs) == 0 {
		// Add buddy.log as combined output file
		if err := router.AddCombinedFile("buddy.log", stripAnsi); err != nil {
			return nil, nil, false, err
		}
		fileMap["buddy.log"] = true

		// Also show on terminal
		router.AddStdoutTerminal()
		router.AddStderrTerminal()

		return router, commandArgs, usePty, nil
	}

	// Process routing arguments
	for _, arg := range filteredArgs {

		arg = strings.ReplaceAll(arg, "2+1", "stderr+stdout")
		arg = strings.ReplaceAll(arg, "1+2", "stdout+stderr")
		arg = strings.ReplaceAll(arg, "2", "stderr")
		arg = strings.ReplaceAll(arg, "1", "stdout")

		if strings.Contains(arg, "=") {
			parts := strings.SplitN(arg, "=", 2)
			streams := parts[0]
			filepath := parts[1]

			_, exists := fileMap[filepath]

			if strings.Contains(streams, "+") {
				if !exists {
					if err := router.AddCombinedFile(filepath, stripAnsi); err != nil {
						return nil, nil, false, err
					}
					fileMap[filepath] = true
				}
			} else if streams == "stdout" {
				if !exists {
					if err := router.AddStdoutFile(filepath, stripAnsi); err != nil {
						return nil, nil, false, err
					}
					fileMap[filepath] = true
				}
			} else if streams == "stderr" {
				if !exists {
					if err := router.AddStderrFile(filepath, stripAnsi); err != nil {
						return nil, nil, false, err
					}
					fileMap[filepath] = true
				}
			}
		} else {
			if strings.Contains(arg, "+") {
				router.AddStdoutTerminal()
				router.AddStderrTerminal()
			} else if arg == "stdout" {
				router.AddStdoutTerminal()
			} else if arg == "stderr" {
				router.AddStderrTerminal()
			}
		}
	}

	return router, commandArgs, usePty, nil
}



func runWithPipes(router *OutputRouter, commandArgs []string) error {
	cmd := exec.Command(commandArgs[0], commandArgs[1:]...)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		return err
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		if cmd.Process != nil {
			cmd.Process.Signal(os.Interrupt)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		buf := make([]byte, 32*1024)
		for {
			n, err := stdoutPipe.Read(buf)
			if n > 0 {
				router.WriteStdout(buf[:n])
			}
			if err != nil {
				break
			}
		}
	}()

	go func() {
		defer wg.Done()
		buf := make([]byte, 32*1024)
		for {
			n, err := stderrPipe.Read(buf)
			if n > 0 {
				router.WriteStderr(buf[:n])
			}
			if err != nil {
				break
			}
		}
	}()

	wg.Wait()
	return cmd.Wait()
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `outputbuddy %s - Flexible output redirection with color preservation

Usage: outputbuddy [options] -- command [args...]

Default behavior (no options):
  Redirects both stdout and stderr to buddy.log AND displays on terminal

Options:
  2=file.log or stderr=file.log          - redirect stderr to file
  1=file.log or stdout=file.log          - redirect stdout to file
  2+1=file.log or stderr+stdout=file.log - redirect both to same file
  2 or stderr                            - show stderr on terminal
  1 or stdout                            - show stdout on terminal
  2+1 or stderr+stdout                   - show both on terminal
  --no-pty                               - disable PTY mode
  --keep-ansi                            - keep ANSI codes in files
  --version, -v                          - show version

Examples:
  ob -- python script.py                 - logs to buddy.log + terminal (default)
  ob 2+1=output.log 2+1 -- python script.py
  ob 2=err.log 1=out.log -- make
  ob 2=err.log 2 -- ./program
`, VERSION)
}

func main() {
	if len(os.Args) >= 2 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("outputbuddy version %s\n", VERSION)
		os.Exit(0)
	}

	if len(os.Args) < 3 {
		printUsage()
		os.Exit(1)
	}

	router, commandArgs, usePty, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		printUsage()
		os.Exit(1)
	}
	defer router.Close()

	var runErr error
	if usePty && term.IsTerminal(int(os.Stdout.Fd())) {
		runErr = runWithPty(router, commandArgs)
	} else {
		runErr = runWithPipes(router, commandArgs)
	}

	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				os.Exit(status.ExitStatus())
			}
		}
		os.Exit(1)
	}

	os.Exit(0)
}