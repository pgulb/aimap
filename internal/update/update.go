package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
)

const repo = "pgulb/aimap"

// Do performs the self-update: checks the latest release on GitHub and replaces
// the running binary if a newer version is available.
func Do(currentVersion string, verbose bool) error {
	if !CanSelfUpdate() {
		return fmt.Errorf("self-update not enabled. Run 'aimap --enable-self-update' to enable it")
	}

	latest, err := fetchLatestTag()
	if err != nil {
		return fmt.Errorf("failed to fetch latest release: %w", err)
	}

	if currentVersion == latest {
		if verbose {
			fmt.Fprintf(os.Stderr, "aimap: already up to date (%s)\n", currentVersion)
		} else {
			fmt.Printf("aimap is already up to date (%s).\n", currentVersion)
		}
		return nil
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "aimap: updating from %s to %s\n", currentVersion, latest)
	}

	currentExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot determine current binary path: %w", err)
	}

	if err := installRelease(latest, currentExe, verbose); err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	fmt.Printf("Updated aimap to %s.\n", latest)
	return nil
}

// installRelease downloads and extracts the release archive, then replaces the binary.
func installRelease(tag, currentExe string, verbose bool) error {
	archiveName := fmt.Sprintf("aimap_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		archiveName = fmt.Sprintf("aimap_%s_%s.zip", runtime.GOOS, runtime.GOARCH)
	}

	url := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", repo, tag, archiveName)

	if verbose {
		fmt.Fprintf(os.Stderr, "aimap: downloading %s\n", url)
	}

	tmpDir, err := os.MkdirTemp("", "aimap-update-*")
	if err != nil {
		return fmt.Errorf("temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	archivePath := tmpDir + "/" + archiveName
	if err := downloadFile(url, archivePath); err != nil {
		return fmt.Errorf("download: %w", err)
	}

	if runtime.GOOS == "windows" {
		return extractZip(archivePath, tmpDir, currentExe)
	}
	return extractTarGz(archivePath, tmpDir, currentExe)
}

// fetchLatestTag queries the GitHub API for the latest release tag.
func fetchLatestTag() (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	resp, err := httpGet(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 403 || resp.StatusCode == 429 {
		return "", fmt.Errorf("GitHub API rate limit exceeded (HTTP %d)", resp.StatusCode)
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("GitHub API returned HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.Unmarshal(body, &release); err != nil {
		return "", err
	}
	if release.TagName == "" {
		return "", fmt.Errorf("no releases found")
	}
	return release.TagName, nil
}

var httpGet = http.Get
