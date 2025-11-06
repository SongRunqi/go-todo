package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/SongRunqi/go-todo/internal/version"
)

const (
	// GitHub repository information
	githubOwner = "SongRunqi"
	githubRepo  = "go-todo"
	githubAPI   = "https://api.github.com"

	// Timeout for HTTP requests
	httpTimeout = 30 * time.Second
)

// Release represents a GitHub release
type Release struct {
	TagName    string  `json:"tag_name"`
	Name       string  `json:"name"`
	Body       string  `json:"body"`
	Draft      bool    `json:"draft"`
	Prerelease bool    `json:"prerelease"`
	CreatedAt  string  `json:"created_at"`
	Assets     []Asset `json:"assets"`
}

// Asset represents a release asset
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// Updater handles application updates
type Updater struct {
	currentVersion string
	httpClient     *http.Client
}

// New creates a new Updater instance
func New() *Updater {
	return &Updater{
		currentVersion: version.Version,
		httpClient: &http.Client{
			Timeout: httpTimeout,
		},
	}
}

// CheckForUpdates checks if a new version is available
func (u *Updater) CheckForUpdates() (*Release, bool, error) {
	release, err := u.getLatestRelease()
	if err != nil {
		return nil, false, fmt.Errorf("failed to check for updates: %w", err)
	}

	// Skip draft and prerelease versions
	if release.Draft || release.Prerelease {
		return nil, false, nil
	}

	// Compare versions
	latestVersion := strings.TrimPrefix(release.TagName, "v")
	currentVersion := strings.TrimPrefix(u.currentVersion, "v")

	// Skip if current version is "dev"
	if currentVersion == "dev" {
		return release, false, nil
	}

	// Simple version comparison
	hasUpdate := latestVersion != currentVersion

	return release, hasUpdate, nil
}

// Update downloads and installs the latest version
func (u *Updater) Update() error {
	release, hasUpdate, err := u.CheckForUpdates()
	if err != nil {
		return err
	}

	if !hasUpdate {
		return fmt.Errorf("already running the latest version (%s)", u.currentVersion)
	}

	// Find the appropriate asset for current platform
	assetName := u.getAssetName()
	checksumAssetName := assetName + ".sha256"

	var binaryAsset, checksumAsset *Asset
	for i := range release.Assets {
		if release.Assets[i].Name == assetName {
			binaryAsset = &release.Assets[i]
		}
		if release.Assets[i].Name == checksumAssetName {
			checksumAsset = &release.Assets[i]
		}
	}

	if binaryAsset == nil {
		return fmt.Errorf("no binary found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	// Download the binary
	binaryData, err := u.downloadFile(binaryAsset.BrowserDownloadURL)
	if err != nil {
		return fmt.Errorf("failed to download binary: %w", err)
	}

	// Download and verify checksum if available
	if checksumAsset != nil {
		checksumData, err := u.downloadFile(checksumAsset.BrowserDownloadURL)
		if err != nil {
			return fmt.Errorf("failed to download checksum: %w", err)
		}

		expectedChecksum := strings.TrimSpace(string(checksumData))
		actualChecksum := u.calculateSHA256(binaryData)

		if actualChecksum != expectedChecksum {
			return fmt.Errorf("checksum verification failed: expected %s, got %s", expectedChecksum, actualChecksum)
		}
	}

	// Replace the current binary
	if err := u.replaceBinary(binaryData); err != nil {
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	return nil
}

// getLatestRelease fetches the latest release from GitHub
func (u *Updater) getLatestRelease() (*Release, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", githubAPI, githubOwner, githubRepo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

// downloadFile downloads a file from the given URL
func (u *Updater) downloadFile(url string) ([]byte, error) {
	resp, err := u.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// calculateSHA256 calculates the SHA256 checksum of data
func (u *Updater) calculateSHA256(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// getAssetName returns the asset name for the current platform
func (u *Updater) getAssetName() string {
	binary := "todo"
	if runtime.GOOS == "windows" {
		binary = "todo.exe"
	}

	return fmt.Sprintf("todo-%s-%s-%s", runtime.GOOS, runtime.GOARCH, binary)
}

// replaceBinary replaces the current binary with the new one
func (u *Updater) replaceBinary(newBinary []byte) error {
	// Get the path of the current executable
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Resolve symlinks
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve symlinks: %w", err)
	}

	// Create backup
	backupPath := execPath + ".backup"
	if err := u.copyFile(execPath, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Write new binary to a temporary file
	tempPath := execPath + ".new"
	if err := os.WriteFile(tempPath, newBinary, 0755); err != nil {
		return fmt.Errorf("failed to write new binary: %w", err)
	}

	// Replace the current binary
	if err := os.Rename(tempPath, execPath); err != nil {
		// Restore backup on failure
		_ = os.Rename(backupPath, execPath)
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	// Remove backup
	_ = os.Remove(backupPath)

	return nil
}

// copyFile copies a file from src to dst
func (u *Updater) copyFile(src, dst string) error {
	sourceData, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, sourceData, sourceInfo.Mode())
}

// GetCurrentVersion returns the current version
func (u *Updater) GetCurrentVersion() string {
	return u.currentVersion
}
