package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

const (
	apiURL     = "https://api.github.com/repos/AgusRdz/logr/releases/latest"
	releaseURL = "https://github.com/AgusRdz/logr/releases/download/%s/%s"
)

// Run checks for a newer release on GitHub and replaces the binary if found.
// Prints status to stdout.
func Run(currentVersion string) {
	latest := CheckLatest()
	if latest == "" {
		fmt.Println("could not determine latest version")
		return
	}
	if latest == currentVersion {
		fmt.Printf("logr is up to date (%s)\n", currentVersion)
		return
	}

	binary := binaryName()
	url := fmt.Sprintf(releaseURL, latest, binary)
	fmt.Printf("downloading %s...\n", url)

	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		fmt.Printf("download failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("download failed: HTTP %d\n", resp.StatusCode)
		return
	}

	execPath, err := os.Executable()
	if err != nil {
		fmt.Printf("could not determine executable path: %v\n", err)
		return
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		fmt.Printf("could not resolve executable path: %v\n", err)
		return
	}

	// Write to temp file
	tmp, err := os.CreateTemp(filepath.Dir(execPath), "logr-update-*")
	if err != nil {
		fmt.Printf("could not create temp file: %v\n", err)
		return
	}
	tmpPath := tmp.Name()

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		fmt.Printf("write failed: %v\n", err)
		return
	}
	tmp.Close()

	// chmod +x on Unix
	if runtime.GOOS != "windows" {
		if err := os.Chmod(tmpPath, 0755); err != nil {
			os.Remove(tmpPath)
			fmt.Printf("chmod failed: %v\n", err)
			return
		}
	}

	// Replace binary
	if err := os.Rename(tmpPath, execPath); err != nil {
		if runtime.GOOS == "windows" {
			// Windows locks the running exe. Move the new binary to a staging path,
			// then launch a detached batch script that swaps the files after we exit.
			newPath := execPath + ".new"
			if renameErr := os.Rename(tmpPath, newPath); renameErr != nil {
				os.Remove(tmpPath)
				fmt.Printf("update failed: %v\n", err)
				return
			}
			if scheduleErr := scheduleWindowsSwap(newPath, execPath); scheduleErr != nil {
				fmt.Printf("updated binary written to %s\n", newPath)
				fmt.Println("Replace the running binary manually to apply the update.")
				return
			}
			fmt.Printf("updated to %s — restart logr to use the new version\n", latest)
			os.Exit(0)
		}
		os.Remove(tmpPath)
		fmt.Printf("update failed: %v\n", err)
		return
	}

	fmt.Printf("updated to %s\n", latest)
}

// scheduleWindowsSwap writes a self-deleting batch script that waits for the
// current process to release the exe lock, then moves newPath over execPath.
func scheduleWindowsSwap(newPath, execPath string) error {
	bat, err := os.CreateTemp(filepath.Dir(execPath), "logr-swap-*.bat")
	if err != nil {
		return err
	}
	batPath := bat.Name()

	// ping -n 2 gives ~1 second delay — enough for our process to exit.
	script := fmt.Sprintf("@echo off\r\nping -n 2 127.0.0.1 > nul\r\nmove /Y %q %q\r\ndel %q\r\n",
		newPath, execPath, batPath)
	if _, err := bat.WriteString(script); err != nil {
		bat.Close()
		os.Remove(batPath)
		return err
	}
	bat.Close()

	// /MIN starts the script in a detached minimized window, independent of
	// the parent process. /B would keep it in the same console and die with us.
	cmd := exec.Command("cmd", "/C", "start", "/MIN", "", batPath)
	return cmd.Start()
}

// CheckLatest returns the latest version tag from GitHub releases API.
// Returns "" on any error.
func CheckLatest() string {
	resp, err := http.Get(apiURL) //nolint:gosec
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	var payload struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return ""
	}
	return payload.TagName
}

// NotifyIfUpdateAvailable prints a dim hint if a newer version exists.
// Non-blocking: spawns a goroutine, returns immediately.
func NotifyIfUpdateAvailable(currentVersion string, w io.Writer) {
	go func() {
		latest := CheckLatest()
		if latest == "" || latest == currentVersion {
			return
		}
		fmt.Fprintf(w, "\033[2mlogr %s is available (current: %s) - run `logr update` to upgrade\033[0m\n",
			latest, currentVersion)
	}()
}

// binaryName returns the expected release asset filename for the current platform.
func binaryName() string {
	name := fmt.Sprintf("logr-%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return name
}
