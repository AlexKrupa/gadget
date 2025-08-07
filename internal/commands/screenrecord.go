package commands

import (
	"fmt"
	"gadget/internal/adb"
	"gadget/internal/config"
	"gadget/internal/logger"
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
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("android-vid-%s.mp4", timestamp)
	localPath := filepath.Join(cfg.MediaPath, filename)

	remoteFilename := fmt.Sprintf("screenrecord_%s.mp4", timestamp)
	remotePath := "/sdcard/" + remoteFilename

	adbPath := cfg.GetADBPath()
	cmd := exec.Command(adbPath, "-s", device.Serial, "shell", "screenrecord", remotePath)

	err := cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start screen recording: %w", err)
	}

	recording := &ScreenRecording{
		Device:     device,
		Cmd:        cmd,
		LocalPath:  localPath,
		Config:     cfg,
		RemotePath: remotePath,
	}

	return recording, nil
}

// StopAndSave stops the recording and saves it to local machine
func (r *ScreenRecording) StopAndSave() error {
	if r.Cmd != nil && r.Cmd.Process != nil {
		err := r.Cmd.Process.Signal(syscall.SIGINT)
		if err != nil {
			return fmt.Errorf("failed to stop recording: %w", err)
		}

		r.Cmd.Wait()
	}

	time.Sleep(2 * time.Second)

	adbPath := r.Config.GetADBPath()
	checkCmd := exec.Command(adbPath, "-s", r.Device.Serial, "shell", "ls", "-la", r.RemotePath)
	checkOutput, checkErr := checkCmd.CombinedOutput()
	if checkErr != nil {
		return fmt.Errorf("recording file not found on device: %s", string(checkOutput))
	}

	logger.Info("File on device: %s", string(checkOutput))

	localDir := filepath.Dir(r.LocalPath)
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return fmt.Errorf("failed to create local directory %s: %w", localDir, err)
	}

	// Try to pull the file from device - try different approaches
	logger.Info("Attempting pull command: %s -s %s pull %s %s", adbPath, r.Device.Serial, r.RemotePath, r.LocalPath)
	pullCmd := exec.Command(adbPath, "-s", r.Device.Serial, "pull", r.RemotePath, r.LocalPath)
	pullCmd.Stderr = nil
	pullCmd.Stdout = nil
	pullOutput, err := pullCmd.CombinedOutput()

	if err != nil {
		logger.Error("Pull attempt 1 failed. Error: %v", err)
		logger.Error("Pull attempt 1 output: %q", string(pullOutput))

		// Second try: without device serial (if only one device)
		pullCmd2 := exec.Command(adbPath, "pull", r.RemotePath, r.LocalPath)
		pullOutput2, err2 := pullCmd2.CombinedOutput()

		if err2 != nil {
			logger.Error("Pull attempt 2 failed. Error: %v", err2)
			logger.Error("Pull attempt 2 output: %q", string(pullOutput2))

			return fmt.Errorf("both pull attempts failed. First: %v (output: %q), Second: %v (output: %q)",
				err, string(pullOutput), err2, string(pullOutput2))
		}
	}

	cleanCmd := exec.Command(adbPath, "-s", r.Device.Serial, "shell", "rm", r.RemotePath)
	cleanCmd.Run() // Ignore cleanup errors

	logger.Success("Screen recording saved to: %s", r.LocalPath)
	return nil
}
