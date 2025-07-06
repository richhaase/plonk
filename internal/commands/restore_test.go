package commands

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRestoreCommand_ListBackups_WithExistingBackups(t *testing.T) {
	// Setup temporary directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempHome)
	
	// Create plonk directory and config
	plonkDir := filepath.Join(tempHome, ".config", "plonk")
	err := os.MkdirAll(plonkDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create plonk directory: %v", err)
	}
	
	// Create config file with backup configuration
	configContent := `settings:
  default_manager: homebrew

backup:
  location: "~/.config/plonk/backups"
  keep_count: 5
`
	
	configPath := filepath.Join(plonkDir, "plonk.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
	
	// Create backup directory with some existing backups
	backupDir := filepath.Join(tempHome, ".config", "plonk", "backups")
	err = os.MkdirAll(backupDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create backup directory: %v", err)
	}
	
	// Create some backup files
	backupFiles := []struct {
		filename string
		content  string
	}{
		{"zshrc.backup.20241206-143022", "# Old zshrc content\nalias ls='ls -la'"},
		{"zshrc.backup.20241206-150330", "# Newer zshrc content\nalias ll='ls -la'"},
		{"vimrc.backup.20241206-143022", "\" Old vimrc\nset number"},
		{"gitconfig.backup.20241205-120000", "[user]\n\tname = Test User"},
	}
	
	for _, backup := range backupFiles {
		backupPath := filepath.Join(backupDir, backup.filename)
		err = os.WriteFile(backupPath, []byte(backup.content), 0644)
		if err != nil {
			t.Fatalf("Failed to create backup file %s: %v", backup.filename, err)
		}
	}
	
	// Test restore --list command
	output, err := runRestoreCommand([]string{"--list"})
	if err != nil {
		t.Fatalf("Restore --list command failed: %v", err)
	}
	
	// Verify output contains all backup timestamps
	if !strings.Contains(output, "20241206-143022") {
		t.Error("Expected output to contain timestamp 20241206-143022")
	}
	if !strings.Contains(output, "20241206-150330") {
		t.Error("Expected output to contain timestamp 20241206-150330")
	}
	if !strings.Contains(output, "20241205-120000") {
		t.Error("Expected output to contain timestamp 20241205-120000")
	}
	
	// Verify output is organized by original file
	if !strings.Contains(output, "~/.zshrc") {
		t.Error("Expected output to show original file path ~/.zshrc")
	}
	if !strings.Contains(output, "~/.vimrc") {
		t.Error("Expected output to show original file path ~/.vimrc")
	}
	if !strings.Contains(output, "~/.gitconfig") {
		t.Error("Expected output to show original file path ~/.gitconfig")
	}
}

func TestRestoreCommand_ListBackups_NoBackups(t *testing.T) {
	// Setup temporary directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempHome)
	
	// Create plonk directory and config
	plonkDir := filepath.Join(tempHome, ".config", "plonk")
	err := os.MkdirAll(plonkDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create plonk directory: %v", err)
	}
	
	// Create config file
	configContent := `settings:
  default_manager: homebrew

backup:
  location: "~/.config/plonk/backups"
  keep_count: 5
`
	
	configPath := filepath.Join(plonkDir, "plonk.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
	
	// Test restore --list command when no backups exist
	output, err := runRestoreCommand([]string{"--list"})
	if err != nil {
		t.Fatalf("Restore --list command should not fail when no backups exist: %v", err)
	}
	
	// Verify output indicates no backups found
	if !strings.Contains(output, "No backups found") {
		t.Error("Expected output to indicate no backups found")
	}
}

func TestRestoreCommand_ListBackups_NoConfig(t *testing.T) {
	// Setup temporary directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempHome)
	
	// Test restore --list command when no config exists
	_, err := runRestoreCommand([]string{"--list"})
	if err == nil {
		t.Error("Expected error when no config file exists")
	}
	
	if !strings.Contains(err.Error(), "config file not found") {
		t.Errorf("Expected error about config file not found, got: %v", err)
	}
}

// runRestoreCommand is a test helper function
func runRestoreCommand(args []string) (string, error) {
	// Capture stdout to get command output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	// Buffer to capture output
	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()
	
	// Reset flags to avoid test interference
	restoreCmd.Flags().Set("list", "false")
	restoreCmd.Flags().Set("all", "false")
	restoreCmd.Flags().Set("timestamp", "")
	
	// Parse the arguments to set flags
	for i, arg := range args {
		switch arg {
		case "--list":
			restoreCmd.Flags().Set("list", "true")
		case "--all":
			restoreCmd.Flags().Set("all", "true")
		case "--timestamp":
			if i+1 < len(args) {
				restoreCmd.Flags().Set("timestamp", args[i+1])
			}
		}
	}
	
	// Filter out flag arguments to get positional args
	var positionalArgs []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "--") {
			positionalArgs = append(positionalArgs, arg)
		} else if arg == "--timestamp" && i+1 < len(args) {
			i++ // Skip the timestamp value
		}
	}
	
	// Execute the command
	err := restoreCmdRun(restoreCmd, positionalArgs)
	
	// Restore stdout and get captured output
	w.Close()
	os.Stdout = oldStdout
	output := <-outC
	
	return output, err
}

func TestRestoreCommand_RestoreLatestFile(t *testing.T) {
	// Setup temporary directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempHome)
	
	// Create existing .zshrc file (will be overwritten by restore)
	existingZshrc := filepath.Join(tempHome, ".zshrc")
	currentContent := "# Current zshrc content\nalias current='echo current'"
	err := os.WriteFile(existingZshrc, []byte(currentContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create existing .zshrc: %v", err)
	}
	
	// Create plonk directory and config
	plonkDir := filepath.Join(tempHome, ".config", "plonk")
	err = os.MkdirAll(plonkDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create plonk directory: %v", err)
	}
	
	// Create config file
	configContent := `settings:
  default_manager: homebrew

backup:
  location: "~/.config/plonk/backups"
  keep_count: 5
`
	
	configPath := filepath.Join(plonkDir, "plonk.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
	
	// Create backup directory with backup files
	backupDir := filepath.Join(tempHome, ".config", "plonk", "backups")
	err = os.MkdirAll(backupDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create backup directory: %v", err)
	}
	
	// Create backup files (older and newer)
	olderBackup := filepath.Join(backupDir, "zshrc.backup.20241205-120000")
	olderContent := "# Older zshrc backup\nalias old='echo old'"
	err = os.WriteFile(olderBackup, []byte(olderContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create older backup: %v", err)
	}
	
	newerBackup := filepath.Join(backupDir, "zshrc.backup.20241206-150330")
	newerContent := "# Newer zshrc backup\nalias new='echo new'"
	err = os.WriteFile(newerBackup, []byte(newerContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create newer backup: %v", err)
	}
	
	// Test restore command for latest backup
	output, err := runRestoreCommand([]string{existingZshrc})
	if err != nil {
		t.Fatalf("Restore command failed: %v", err)
	}
	
	// Verify output indicates successful restore
	if !strings.Contains(output, "Restored") {
		t.Error("Expected output to indicate successful restore")
	}
	
	// Verify file was restored with latest backup content
	restoredContent, err := os.ReadFile(existingZshrc)
	if err != nil {
		t.Fatalf("Failed to read restored file: %v", err)
	}
	
	if string(restoredContent) != newerContent {
		t.Errorf("File was not restored with latest backup content.\nExpected: %s\nGot: %s", 
			newerContent, string(restoredContent))
	}
}

func TestRestoreCommand_RestoreFile_NoBackups(t *testing.T) {
	// Setup temporary directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempHome)
	
	// Create plonk directory and config
	plonkDir := filepath.Join(tempHome, ".config", "plonk")
	err := os.MkdirAll(plonkDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create plonk directory: %v", err)
	}
	
	// Create config file
	configContent := `settings:
  default_manager: homebrew

backup:
  location: "~/.config/plonk/backups"
  keep_count: 5
`
	
	configPath := filepath.Join(plonkDir, "plonk.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
	
	// Test restore command when no backups exist
	_, err = runRestoreCommand([]string{filepath.Join(tempHome, ".zshrc")})
	if err == nil {
		t.Error("Expected error when trying to restore file with no backups")
	}
	
	if !strings.Contains(err.Error(), "no backups found") {
		t.Errorf("Expected error about no backups found, got: %v", err)
	}
}

func TestRestoreCommand_RestoreFile_FileNotExist(t *testing.T) {
	// Setup temporary directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempHome)
	
	// Create plonk directory and config
	plonkDir := filepath.Join(tempHome, ".config", "plonk")
	err := os.MkdirAll(plonkDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create plonk directory: %v", err)
	}
	
	// Create config file
	configContent := `settings:
  default_manager: homebrew

backup:
  location: "~/.config/plonk/backups"
  keep_count: 5
`
	
	configPath := filepath.Join(plonkDir, "plonk.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
	
	// Create backup directory with backup for different file
	backupDir := filepath.Join(tempHome, ".config", "plonk", "backups")
	err = os.MkdirAll(backupDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create backup directory: %v", err)
	}
	
	// Create backup for vimrc but try to restore zshrc
	vimrcBackup := filepath.Join(backupDir, "vimrc.backup.20241206-150330")
	err = os.WriteFile(vimrcBackup, []byte("\" vimrc backup"), 0644)
	if err != nil {
		t.Fatalf("Failed to create vimrc backup: %v", err)
	}
	
	// Test restore command for file that has no backups
	_, err = runRestoreCommand([]string{filepath.Join(tempHome, ".zshrc")})
	if err == nil {
		t.Error("Expected error when trying to restore file with no backups")
	}
	
	if !strings.Contains(err.Error(), "no backups found") {
		t.Errorf("Expected error about no backups found, got: %v", err)
	}
}

func TestRestoreCommand_RestoreSpecificTimestamp(t *testing.T) {
	// Setup temporary directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempHome)
	
	// Create existing .zshrc file (will be overwritten by restore)
	existingZshrc := filepath.Join(tempHome, ".zshrc")
	currentContent := "# Current zshrc content\nalias current='echo current'"
	err := os.WriteFile(existingZshrc, []byte(currentContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create existing .zshrc: %v", err)
	}
	
	// Create plonk directory and config
	plonkDir := filepath.Join(tempHome, ".config", "plonk")
	err = os.MkdirAll(plonkDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create plonk directory: %v", err)
	}
	
	// Create config file
	configContent := `settings:
  default_manager: homebrew

backup:
  location: "~/.config/plonk/backups"
  keep_count: 5
`
	
	configPath := filepath.Join(plonkDir, "plonk.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
	
	// Create backup directory with backup files
	backupDir := filepath.Join(tempHome, ".config", "plonk", "backups")
	err = os.MkdirAll(backupDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create backup directory: %v", err)
	}
	
	// Create backup files with different timestamps
	olderBackup := filepath.Join(backupDir, "zshrc.backup.20241205-120000")
	olderContent := "# Older zshrc backup\nalias old='echo old'"
	err = os.WriteFile(olderBackup, []byte(olderContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create older backup: %v", err)
	}
	
	newerBackup := filepath.Join(backupDir, "zshrc.backup.20241206-150330")
	newerContent := "# Newer zshrc backup\nalias new='echo new'"
	err = os.WriteFile(newerBackup, []byte(newerContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create newer backup: %v", err)
	}
	
	// Test restore command with specific older timestamp
	output, err := runRestoreCommand([]string{existingZshrc, "--timestamp", "20241205-120000"})
	if err != nil {
		t.Fatalf("Restore command with timestamp failed: %v", err)
	}
	
	// Verify output indicates successful restore with correct timestamp
	if !strings.Contains(output, "Restored") {
		t.Error("Expected output to indicate successful restore")
	}
	if !strings.Contains(output, "20241205-120000") {
		t.Error("Expected output to show the specific timestamp used")
	}
	
	// Verify file was restored with older backup content (not the newer one)
	restoredContent, err := os.ReadFile(existingZshrc)
	if err != nil {
		t.Fatalf("Failed to read restored file: %v", err)
	}
	
	if string(restoredContent) != olderContent {
		t.Errorf("File was not restored with specific timestamp backup content.\nExpected: %s\nGot: %s", 
			olderContent, string(restoredContent))
	}
}

func TestRestoreCommand_RestoreInvalidTimestamp(t *testing.T) {
	// Setup temporary directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempHome)
	
	// Create plonk directory and config
	plonkDir := filepath.Join(tempHome, ".config", "plonk")
	err := os.MkdirAll(plonkDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create plonk directory: %v", err)
	}
	
	// Create config file
	configContent := `settings:
  default_manager: homebrew

backup:
  location: "~/.config/plonk/backups"
  keep_count: 5
`
	
	configPath := filepath.Join(plonkDir, "plonk.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
	
	// Create backup directory with one backup
	backupDir := filepath.Join(tempHome, ".config", "plonk", "backups")
	err = os.MkdirAll(backupDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create backup directory: %v", err)
	}
	
	// Create a backup file
	backup := filepath.Join(backupDir, "zshrc.backup.20241206-150330")
	err = os.WriteFile(backup, []byte("# backup content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}
	
	// Test restore command with non-existent timestamp
	_, err = runRestoreCommand([]string{filepath.Join(tempHome, ".zshrc"), "--timestamp", "20241201-000000"})
	if err == nil {
		t.Error("Expected error when trying to restore with non-existent timestamp")
	}
	
	if !strings.Contains(err.Error(), "backup with timestamp") || !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected error about timestamp not found, got: %v", err)
	}
}