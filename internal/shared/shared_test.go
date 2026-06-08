package shared

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestHasBin(t *testing.T) {
	// "ls" should exist on all Unix systems
	if !HasBin("ls") {
		t.Error("expected 'ls' to be found in PATH")
	}
	if HasBin("nonexistent_command_xyz_12345") {
		t.Error("expected nonexistent command to not be found")
	}
}

func TestFileExists(t *testing.T) {
	tmp := t.TempDir()
	testFile := filepath.Join(tmp, "exists.txt")
	os.WriteFile(testFile, []byte("test"), 0644)

	if !FileExists(testFile) {
		t.Error("expected file to exist")
	}
	if FileExists(filepath.Join(tmp, "nonexistent.txt")) {
		t.Error("expected file to not exist")
	}
}

func TestRunInstall(t *testing.T) {
	err := RunInstall("echo hello")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	err = RunInstall("false")
	if err == nil {
		t.Error("expected error for 'false' command")
	}
}

func TestSafeReadFile_InGitRepo(t *testing.T) {
	// Create a temp git repo
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(origDir)

	exec.Command("git", "init").Run()
	os.WriteFile(filepath.Join(tmp, "test.txt"), []byte("hello"), 0644)

	data, err := SafeReadFile("test.txt")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if string(data) != "hello" {
		t.Errorf("expected 'hello', got '%s'", string(data))
	}
}

func TestSafeReadFile_LargeFileBlocked(t *testing.T) {
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(origDir)

	exec.Command("git", "init").Run()
	bigFile := filepath.Join(tmp, "big.txt")
	data := make([]byte, 2*1024*1024) // 2MB
	os.WriteFile(bigFile, data, 0644)

	_, err := SafeReadFile("big.txt")
	if err == nil {
		t.Error("expected error for file > 1MB")
	}
}
