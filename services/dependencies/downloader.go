package dependencies

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var httpClient = &http.Client{
	Timeout: 300 * time.Second,
}

type Downloader struct {
	srcDir string
}

func NewDownloader() *Downloader {
	return &Downloader{
		srcDir: GetSrcDir(),
	}
}

func (d *Downloader) Download(dep *Dependency) error {
	if err := EnsureDir(d.srcDir); err != nil {
		return fmt.Errorf("failed to create src directory: %w", err)
	}

	url := dep.GetDownloadURL()
	if url == "" {
		return fmt.Errorf("no download URL for %s", dep.Name)
	}

	fmt.Printf("Downloading %s from %s...\n", dep.DisplayName, url)

	data, err := d.downloadFile(url)
	if err != nil {
		return fmt.Errorf("failed to download %s: %w", dep.Name, err)
	}

	localPath := GetLocalPath(dep.Name)

	if dep.Extract {
		if err := d.extractArchive(data, localPath); err != nil {
			return fmt.Errorf("failed to extract %s: %w", dep.Name, err)
		}
	} else {
		if err := os.WriteFile(localPath, data, 0755); err != nil {
			return fmt.Errorf("failed to write %s: %w", dep.Name, err)
		}
	}

	if err := MakeExecutable(localPath); err != nil {
		return fmt.Errorf("failed to make %s executable: %w", dep.Name, err)
	}

	fmt.Printf("%s downloaded to %s\n", dep.DisplayName, localPath)
	return nil
}

func (d *Downloader) downloadFile(url string) ([]byte, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (d *Downloader) extractArchive(data []byte, localPath string) error {
	contentType := http.DetectContentType(data)
	localDir := filepath.Dir(localPath)

	switch {
	case strings.HasSuffix(localPath, ".zip") || contentType == "application/zip":
		return d.extractZip(data, localDir)
	case strings.HasSuffix(localPath, ".tar.xz") || contentType == "application/x-xz":
		return d.extractTarXZ(data, localDir)
	case strings.HasSuffix(localPath, ".tar.gz"):
		return d.extractTarGZ(data, localDir)
	case contentType == "application/x-sh":
		return d.extractShell(data, localPath)
	case contentType == "application/x-powershell-script":
		return d.extractPowerShell(data, localPath)
	default:
		if strings.Contains(contentType, "text") {
			return d.extractShell(data, localPath)
		}
		return fmt.Errorf("unknown archive type: %s", contentType)
	}
}

func (d *Downloader) extractZip(data []byte, destDir string) error {
	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return err
	}

	binName := "ffmpeg"
	if GetPlatform() == PlatformWindows {
		binName = "ffmpeg.exe"
	}

	for _, f := range zipReader.File {
		if !strings.Contains(f.Name, binName) && !strings.HasSuffix(f.Name, "/") {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		outPath := filepath.Join(destDir, filepath.Base(f.Name))
		outFile, err := os.Create(outPath)
		if err != nil {
			return err
		}
		defer outFile.Close()

		if _, err := io.Copy(outFile, rc); err != nil {
			return err
		}
	}

	return nil
}

func (d *Downloader) extractTarXZ(data []byte, destDir string) error {
	xzReader, err := decompressXZ(data)
	if err != nil {
		return err
	}
	defer xzReader.Close()

	tarReader := tar.NewReader(xzReader)
	binName := "ffmpeg"

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if !strings.Contains(header.Name, binName) {
			continue
		}

		outPath := filepath.Join(destDir, filepath.Base(header.Name))
		outFile, err := os.Create(outPath)
		if err != nil {
			return err
		}
		defer outFile.Close()

		if _, err := io.Copy(outFile, tarReader); err != nil {
			return err
		}
	}

	return nil
}

func (d *Downloader) extractTarGZ(data []byte, destDir string) error {
	return fmt.Errorf("tar.gz extraction not implemented")
}

func (d *Downloader) extractShell(data []byte, localPath string) error {
	tmpFile := localPath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0755); err != nil {
		return err
	}
	defer os.Remove(tmpFile)

	cmd := exec.Command("sh", tmpFile)
	cmd.Env = append(os.Environ(), "DENO_INSTALL="+filepath.Dir(localPath))
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Shell script output: %s\n", output)
		return err
	}

	return nil
}

func (d *Downloader) extractPowerShell(data []byte, localPath string) error {
	tmpFile := localPath + ".ps1"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return err
	}
	defer os.Remove(tmpFile)

	cmd := exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-File", tmpFile)
	cmd.Env = append(os.Environ(), "DENO_INSTALL="+filepath.Dir(localPath))
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("PowerShell output: %s\n", output)
		return err
	}

	return nil
}

func decompressXZ(data []byte) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(data)), nil
}
