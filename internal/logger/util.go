package logger

import (
	"os"
	"time"
)

// formatTime returns the current time formatted as "2006-01-02 15:04:05.000".
func formatTime() string {
	return time.Now().Format("2006-01-02 15:04:05.000")
}

// isTerminal checks if the given file descriptor is a terminal.
func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
