package commands

import (
	"adx/internal/adb"
	"adx/internal/config"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

// ScreenRecording represents an active screen recording session
type ScreenRecording struct {
	Device     adb.Device
	Cmd        *exec.Cmd
	LocalPath  string
	RemotePath string
	Config     *config.Config
}

// StartScreenRecord starts recording the screen using raw ADB
func StartScreenRecord(cfg *config.Config, device adb.Device) (*ScreenRecording, error) {
	// Generate timestamp filename (format: 2021-08-31_14-23-45)
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("android-vid-%s.mp4", timestamp)
	localPath := filepath.Join(cfg.MediaPath, filename)
	
	// Generate unique remote filename to avoid conflicts
	remoteFilename := fmt.Sprintf("screenrecord_%s.mp4", timestamp)
	remotePath := "/sdcard/" + remoteFilename
	
	// Use raw ADB for screen recording
	adbPath := cfg.GetADBPath()
	cmd := exec.Command(adbPath, "-s", device.Serial, "shell", "screenrecord", remotePath)
	
	err := cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start screen recording: %w", err)
	}
	
	recording := &ScreenRecording{
		Device:    device,
		Cmd:       cmd,
		LocalPath: localPath,
		Config:    cfg,
		RemotePath: remotePath,
	}
	
	return recording, nil
}

// StopAndSave stops the recording and saves it to local machine
func (r *ScreenRecording) StopAndSave() error {
	// Stop the recording process gracefully with SIGINT
	if r.Cmd != nil && r.Cmd.Process != nil {
		err := r.Cmd.Process.Signal(syscall.SIGINT)
		if err != nil {
			return fmt.Errorf("failed to stop recording: %w", err)
		}
		
		// Wait for process to finish
		r.Cmd.Wait()
	}
	
	// Wait a moment for file to be finalized
	time.Sleep(2 * time.Second)
	
	// Check if file exists and get details
	adbPath := r.Config.GetADBPath()
	checkCmd := exec.Command(adbPath, "-s", r.Device.Serial, "shell", "ls", "-la", r.RemotePath)
	checkOutput, checkErr := checkCmd.CombinedOutput()
	if checkErr != nil {
		return fmt.Errorf("recording file not found on device: %s", string(checkOutput))
	}
	
	fmt.Printf("File on device: %s\n", string(checkOutput))
	
	// Ensure local directory exists
	localDir := filepath.Dir(r.LocalPath)
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return fmt.Errorf("failed to create local directory %s: %w", localDir, err)
	}
	
	// Try to pull the file from device - try different approaches
	
	// First try: with device serial
	fmt.Printf("Attempting pull command: %s -s %s pull %s %s\n", adbPath, r.Device.Serial, r.RemotePath, r.LocalPath)
	pullCmd := exec.Command(adbPath, "-s", r.Device.Serial, "pull", r.RemotePath, r.LocalPath)
	pullCmd.Stderr = nil
	pullCmd.Stdout = nil
	pullOutput, err := pullCmd.CombinedOutput()
	
	if err != nil {
		// Log the full error details
		fmt.Printf("Pull attempt 1 failed. Error: %v\n", err)
		fmt.Printf("Pull attempt 1 output: %q\n", string(pullOutput))
		
		// Second try: without device serial (if only one device)
		pullCmd2 := exec.Command(adbPath, "pull", r.RemotePath, r.LocalPath)
		pullOutput2, err2 := pullCmd2.CombinedOutput()
		
		if err2 != nil {
			fmt.Printf("Pull attempt 2 failed. Error: %v\n", err2)
			fmt.Printf("Pull attempt 2 output: %q\n", string(pullOutput2))
			
			return fmt.Errorf("both pull attempts failed. First: %v (output: %q), Second: %v (output: %q)", 
				err, string(pullOutput), err2, string(pullOutput2))
		}
	}
	
	// Clean up the remote file
	cleanCmd := exec.Command(adbPath, "-s", r.Device.Serial, "shell", "rm", r.RemotePath)
	cleanCmd.Run() // Ignore cleanup errors
	
	fmt.Printf("Screen recording saved to: %s\n", r.LocalPath)
	return nil
}