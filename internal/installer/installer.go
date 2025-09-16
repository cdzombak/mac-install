package installer

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type Installer struct {
	workDir string
}

func New(workDir string) *Installer {
	return &Installer{
		workDir: workDir,
	}
}

func (i *Installer) Install(installSteps []map[string]string, artifactPath string) error {
	for _, step := range installSteps {
		// Check for archive installation which requires special handling
		if archiveURL, hasArchive := step["archive"]; hasArchive {
			fileName, hasFile := step["file"]
			if err := i.installFromArchive(archiveURL, fileName, hasFile, artifactPath); err != nil {
				return fmt.Errorf("archive installation failed: %w", err)
			}
			continue
		}

		// Check for download installation which requires special handling
		if downloadURL, hasDL := step["dl"]; hasDL {
			// Ensure parent directory exists
			if err := os.MkdirAll(filepath.Dir(artifactPath), 0755); err != nil {
				return fmt.Errorf("failed to create directory for download: %w", err)
			}
			if _, err := i.downloadFile(downloadURL, artifactPath); err != nil {
				return fmt.Errorf("download installation failed: %w", err)
			}
			continue
		}

		// Handle regular installation methods
		for method, value := range step {
			if err := i.executeInstallStep(method, value); err != nil {
				return fmt.Errorf("installation step %s %s failed: %w", method, value, err)
			}
		}
	}
	return nil
}

func (i *Installer) Configure(configSteps []map[string]string) error {
	ignoreErrors := false

	for _, step := range configSteps {
		for method, value := range step {
			if method == "ignore_errors" {
				ignoreErrors = strings.ToLower(value) == "true"
				continue
			}

			if err := i.executeConfigStep(method, value); err != nil {
				if ignoreErrors {
					fmt.Printf("Warning: configuration step %s failed (ignored): %v\n", method, err)
					continue
				}
				return fmt.Errorf("configuration step %s failed: %w", method, err)
			}
		}
	}
	return nil
}

func (i *Installer) executeInstallStep(method, value string) error {
	switch method {
	case "brew":
		return i.runCommand("brew", "install", value)
	case "cask":
		return i.runCommand("brew", "install", "--cask", value)
	case "mas":
		appID := i.extractAppStoreID(value)
		return i.runCommand("mas", "install", appID)
	case "npm":
		return i.runCommand("/opt/homebrew/bin/npm", "install", "-g", value)
	case "gem":
		return i.runCommand("brew", "gem", "install", value)
	case "gomod":
		return i.runCommand("brew", "gomod", value)
	case "pipx":
		return i.runCommand("/opt/homebrew/bin/pipx", "install", value)
	case "run":
		return i.runShellCommand(value)
	case "script":
		return i.runScript(value)
	case "archive":
		return fmt.Errorf("archive installation requires special handling with 'file' parameter")
	default:
		return fmt.Errorf("unknown installation method: %s", method)
	}
}

func (i *Installer) executeConfigStep(method, value string) error {
	switch method {
	case "run":
		return i.runShellCommand(value)
	case "script":
		return i.runScript(value)
	default:
		return fmt.Errorf("unknown configuration method: %s", method)
	}
}

func (i *Installer) runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (i *Installer) runShellCommand(command string) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = i.workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (i *Installer) runScript(scriptPath string) error {
	cmd := exec.Command("sh", scriptPath)
	cmd.Dir = i.workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// extractAppStoreID extracts the app ID from either a raw ID or an App Store URL
func (i *Installer) extractAppStoreID(value string) string {
	// If it's already just a number, return it as-is
	if regexp.MustCompile(`^\d+$`).MatchString(value) {
		return value
	}

	// Try to extract ID from App Store URL
	// URLs are typically: https://apps.apple.com/us/app/app-name/id123456789?mt=12
	// We want to extract the number after "id"
	re := regexp.MustCompile(`/id(\d+)`)
	matches := re.FindStringSubmatch(value)
	if len(matches) > 1 {
		return matches[1]
	}

	// If we can't parse it, return the original value and let mas handle the error
	return value
}

func (i *Installer) ArtifactExists(artifactPath string) bool {
	// If the path contains asterisks, treat it as a wildcard pattern
	if strings.Contains(artifactPath, "*") {
		matches, err := filepath.Glob(artifactPath)
		if err != nil {
			return false
		}
		// Return true if at least one match is found
		return len(matches) > 0
	}

	// For non-wildcard paths, use the original logic
	_, err := os.Stat(artifactPath)
	return err == nil
}

func (i *Installer) GetBrewCaveats(packageName string) (string, error) {
	cmd := exec.Command("brew", "caveats", packageName)
	output, err := cmd.Output()
	if err != nil {
		return "", nil
	}

	caveats := strings.TrimSpace(string(output))
	if caveats == "" || strings.Contains(caveats, "has no caveats") {
		return "", nil
	}

	return caveats, nil
}

func (i *Installer) installFromArchive(archiveURL, fileName string, hasFile bool, artifactPath string) error {
	// Create temporary directory for extraction
	tempDir, err := os.MkdirTemp("", "mac-install-archive-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to remove temp directory: %v\n", err)
		}
	}()

	// Download the archive
	archivePath := filepath.Join(tempDir, "archive")
	actualArchivePath, err := i.downloadFile(archiveURL, archivePath)
	if err != nil {
		return fmt.Errorf("failed to download archive: %w", err)
	}

	// Determine archive type and extract/mount
	extractDir := filepath.Join(tempDir, "extracted")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		return fmt.Errorf("failed to create extraction directory: %w", err)
	}

	if err := i.extractArchive(actualArchivePath, extractDir, archiveURL); err != nil {
		return fmt.Errorf("failed to extract archive: %w", err)
	}

	if hasFile {
		// Find the specified file in the extracted contents
		sourcePath, err := i.findFileInDirectory(extractDir, fileName)
		if err != nil {
			return fmt.Errorf("failed to find file '%s' in archive: %w", fileName, err)
		}

		// Determine destination path (assume /Applications for .app files)
		destPath := filepath.Join("/Applications", fileName)

		// Copy the file/directory to the destination
		if err := i.copyFileOrDirectory(sourcePath, destPath); err != nil {
			return fmt.Errorf("failed to copy '%s' to '%s': %w", sourcePath, destPath, err)
		}
	} else {
		// Extract all files to the directory containing the artifact
		destDir := filepath.Dir(artifactPath)

		// Create destination directory if it doesn't exist
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return fmt.Errorf("failed to create destination directory '%s': %w", destDir, err)
		}

		// Copy all files from extraction directory to destination
		if err := i.copyDirectoryContents(extractDir, destDir); err != nil {
			return fmt.Errorf("failed to copy archive contents to '%s': %w", destDir, err)
		}
	}

	return nil
}

func (i *Installer) downloadFile(url, filepath string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status: %s", resp.Status)
	}

	// Determine the correct file extension based on Content-Type or Content-Disposition
	actualFilepath := i.determineFilepath(filepath, resp)

	out, err := os.Create(actualFilepath)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := out.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close output file: %v\n", err)
		}
	}()

	_, err = io.Copy(out, resp.Body)
	return actualFilepath, err
}

func (i *Installer) determineFilepath(originalPath string, resp *http.Response) string {
	// Check Content-Disposition header first for filename
	if cd := resp.Header.Get("Content-Disposition"); cd != "" {
		if matches := regexp.MustCompile(`filename="([^"]+)"`).FindStringSubmatch(cd); len(matches) > 1 {
			return filepath.Join(filepath.Dir(originalPath), matches[1])
		}
		if matches := regexp.MustCompile(`filename=([^;\\s]+)`).FindStringSubmatch(cd); len(matches) > 1 {
			return filepath.Join(filepath.Dir(originalPath), matches[1])
		}
	}

	// Check Content-Type header to determine extension
	contentType := resp.Header.Get("Content-Type")
	var ext string
	switch {
	case strings.Contains(contentType, "application/zip"):
		ext = ".zip"
	case strings.Contains(contentType, "application/x-apple-diskimage"):
		ext = ".dmg"
	case strings.Contains(contentType, "application/gzip"), strings.Contains(contentType, "application/x-gzip"):
		ext = ".tar.gz"
	case strings.Contains(contentType, "application/x-tar"):
		ext = ".tar"
	case strings.Contains(contentType, "application/octet-stream"):
		// For octet-stream, try to guess from the final URL after redirects
		if finalURL := resp.Request.URL.String(); finalURL != "" {
			if strings.Contains(strings.ToLower(finalURL), ".zip") {
				ext = ".zip"
			} else if strings.Contains(strings.ToLower(finalURL), ".dmg") {
				ext = ".dmg"
			} else if strings.Contains(strings.ToLower(finalURL), ".tar.gz") || strings.Contains(strings.ToLower(finalURL), ".tgz") {
				ext = ".tar.gz"
			}
		}
	}

	if ext != "" {
		base := filepath.Base(originalPath)
		if !strings.Contains(base, ".") {
			return originalPath + ext
		}
	}

	return originalPath
}

func (i *Installer) extractArchive(archivePath, extractDir, originalURL string) error {
	// Determine file type from the original URL, fallback to local file path
	lowerURL := strings.ToLower(originalURL)
	lowerPath := strings.ToLower(archivePath)

	isDMG := strings.Contains(lowerURL, ".dmg") || strings.HasSuffix(lowerPath, ".dmg")
	isZIP := strings.Contains(lowerURL, ".zip") || strings.HasSuffix(lowerPath, ".zip")
	isTAR := strings.Contains(lowerURL, ".tar.gz") || strings.Contains(lowerURL, ".tgz") ||
		strings.HasSuffix(lowerPath, ".tar.gz") || strings.HasSuffix(lowerPath, ".tgz")

	if isDMG {
		// Mount DMG and copy contents
		mountPoint := filepath.Join(filepath.Dir(extractDir), "dmg-mount")
		if err := os.MkdirAll(mountPoint, 0755); err != nil {
			return err
		}
		defer func() {
			_ = i.runCommand("hdiutil", "detach", mountPoint)
			_ = os.RemoveAll(mountPoint)
		}()

		if err := i.runCommand("hdiutil", "attach", "-mountpoint", mountPoint, "-nobrowse", "-quiet", archivePath); err != nil {
			return err
		}

		return i.runCommand("cp", "-R", mountPoint+"/.", extractDir)
	} else if isZIP {
		return i.runCommand("unzip", "-q", archivePath, "-d", extractDir)
	} else if isTAR {
		return i.runCommand("tar", "-xzf", archivePath, "-C", extractDir)
	} else {
		return fmt.Errorf("unsupported archive format: unable to determine type from URL '%s' or file '%s'", originalURL, archivePath)
	}
}

func (i *Installer) findFileInDirectory(dir, fileName string) (string, error) {
	var foundPath string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Base(path) == fileName {
			foundPath = path
			return filepath.SkipDir // Stop walking once found
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	if foundPath == "" {
		return "", fmt.Errorf("file not found")
	}

	return foundPath, nil
}

func (i *Installer) copyFileOrDirectory(src, dest string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if srcInfo.IsDir() {
		return i.runCommand("cp", "-R", src, dest)
	} else {
		return i.runCommand("cp", src, dest)
	}
}

func (i *Installer) copyDirectoryContents(src, dest string) error {
	// Use cp to copy all contents of src directory to dest directory
	// The /. syntax copies contents of the source directory, not the directory itself
	return i.runCommand("cp", "-R", src+"/.", dest)
}
