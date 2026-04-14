package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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
			newPath := execPath + ".new"
			if renameErr := os.Rename(tmpPath, newPath); renameErr != nil {
				os.Remove(tmpPath)
				fmt.Printf("update failed: %v\n", err)
				return
			}
			fmt.Printf("updated binary written to %s\n", newPath)
			fmt.Println("Replace the running binary manually or restart to apply the update.")
			return
		}
		os.Remove(tmpPath)
		fmt.Printf("update failed: %v\n", err)
		return
	}

	fmt.Printf("updated to %s\n", latest)
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
