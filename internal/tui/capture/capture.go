package capture

import (
	"bufio"
	"io"
	"os"
	"sync"
)

// OutputCapture captures stdout and stderr during command execution
type OutputCapture struct {
	originalStdout *os.File
	originalStderr *os.File
	stdoutReader   *os.File
	stdoutWriter   *os.File
	stderrReader   *os.File
	stderrWriter   *os.File
	outputCh       chan string
	errorCh        chan string
	wg             sync.WaitGroup
	capturing      bool
	mu             sync.Mutex
}

// NewOutputCapture creates a new output capture instance
func NewOutputCapture() *OutputCapture {
	return &OutputCapture{
		outputCh: make(chan string, 100),
		errorCh:  make(chan string, 100),
	}
}

// Start begins capturing stdout and stderr
func (oc *OutputCapture) Start() error {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	if oc.capturing {
		return nil // Already capturing
	}

	// Save original stdout and stderr
	oc.originalStdout = os.Stdout
	oc.originalStderr = os.Stderr

	// Create pipes for stdout
	stdoutReader, stdoutWriter, err := os.Pipe()
	if err != nil {
		return err
	}
	oc.stdoutReader = stdoutReader
	oc.stdoutWriter = stdoutWriter

	// Create pipes for stderr
	stderrReader, stderrWriter, err := os.Pipe()
	if err != nil {
		stdoutReader.Close()
		stdoutWriter.Close()
		return err
	}
	oc.stderrReader = stderrReader
	oc.stderrWriter = stderrWriter

	// Replace stdout and stderr
	os.Stdout = stdoutWriter
	os.Stderr = stderrWriter

	// Start goroutines to read from pipes
	oc.wg.Add(2)
	go oc.readOutput(stdoutReader, oc.outputCh)
	go oc.readOutput(stderrReader, oc.errorCh)

	oc.capturing = true
	return nil
}

// Stop stops capturing and restores original stdout/stderr
func (oc *OutputCapture) Stop() error {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	if !oc.capturing {
		return nil // Not capturing
	}

	// Restore original stdout and stderr
	os.Stdout = oc.originalStdout
	os.Stderr = oc.originalStderr

	// Close write ends of pipes
	oc.stdoutWriter.Close()
	oc.stderrWriter.Close()

	// Wait for readers to finish
	oc.wg.Wait()

	// Close read ends of pipes
	oc.stdoutReader.Close()
	oc.stderrReader.Close()

	oc.capturing = false
	return nil
}

// readOutput reads from a pipe and sends lines to a channel
func (oc *OutputCapture) readOutput(reader io.Reader, ch chan<- string) {
	defer oc.wg.Done()

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			select {
			case ch <- line:
			default:
				// Channel is full, skip this line
			}
		}
	}
}

// GetAllOutput returns all captured output and errors
func (oc *OutputCapture) GetAllOutput() []string {
	var lines []string

	// Get all stdout
	for {
		select {
		case line := <-oc.outputCh:
			lines = append(lines, line)
		default:
			goto getErrors
		}
	}

getErrors:
	// Get all stderr
	for {
		select {
		case line := <-oc.errorCh:
			lines = append(lines, line)
		default:
			return lines
		}
	}
}

// CaptureFunction executes a function while capturing its output
func CaptureFunction(fn func() error) ([]string, error) {
	capture := NewOutputCapture()

	err := capture.Start()
	if err != nil {
		return nil, err
	}

	// Execute the function
	fnErr := fn()

	// Stop capturing
	stopErr := capture.Stop()
	if stopErr != nil {
		return nil, stopErr
	}

	// Get all captured output
	output := capture.GetAllOutput()

	return output, fnErr
}

// CaptureCommand is a convenience function to capture output from a command function
func CaptureCommand(cmdFunc func() error) (output []string, err error) {
	return CaptureFunction(cmdFunc)
}
