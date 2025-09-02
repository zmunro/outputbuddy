package main

import "bytes"

// Custom scanner that handles both lines and progress updates
func scanLinesOrProgress(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		return i + 1, data[0:i], nil
	}

	if i := bytes.IndexByte(data, '\r'); i >= 0 {
		return i + 1, data[0:i], nil
	}

	if atEOF && len(data) > 0 {
		return len(data), data, nil
	}

	return 0, nil, nil
}
