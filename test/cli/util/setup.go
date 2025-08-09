package util

import (
	"bytes"
	"gadget/internal/logger"
	"io"
	"os"
)

// CaptureLogOutput captures logger output during test execution
func CaptureLogOutput(fn func()) string {
	// Redirect stdout to capture logger output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Set up CLI renderer for testing
	logger.SetRenderer(logger.NewCLIRenderer())

	// Run the function
	fn()

	// Close writer and restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read the captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}
