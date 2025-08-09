package util

import (
	"bytes"
	"fmt"
	"gadget/internal/logger"
	"io"
	"os"
	"strconv"
	"testing"
)

// TestHelperProcess is the test helper that gets executed when we mock commands
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_TEST_HELPER_PROCESS") != "1" {
		return
	}

	// Get the command being run from the args
	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}

	if len(args) == 0 {
		os.Exit(1)
	}

	// Output the mocked response
	stdout := os.Getenv("TEST_STDOUT")
	stderr := os.Getenv("TEST_STDERR")
	exitCodeStr := os.Getenv("TEST_EXIT_CODE")

	if stdout != "" {
		fmt.Fprint(os.Stdout, stdout)
	}
	if stderr != "" {
		fmt.Fprint(os.Stderr, stderr)
	}

	exitCode := 0
	if exitCodeStr != "" {
		if code, err := strconv.Atoi(exitCodeStr); err == nil {
			exitCode = code
		}
	}

	os.Exit(exitCode)
}

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
