package dependencies

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type Platform string

const (
	PlatformLinux   Platform = "linux"
	PlatformWindows Platform = "windows"
	PlatformDarwin  Platform = "darwin"
)

var platform Platform
var srcDir string

func init() {
	switch runtime.GOOS {
	case "linux":
		platform = PlatformLinux
	case "windows":
		platform = PlatformWindows
	case "darwin":
		platform = PlatformDarwin
	}

	execPath, err := os.Executable()
	if err != nil {
		srcDir = "./src"
	} else {
		dir := filepath.Dir(execPath)
		srcDir = filepath.Join(dir, "src")
	}
}

func GetPlatform() Platform {
	return platform
}

func GetSrcDir() string {
	return srcDir
}

func FindCommand(name string) (found bool, path string) {
	localSrc := GetSrcDir()
	exts := getExtensions()

	if found, path := checkPath(name, localSrc, exts); found {
		return true, path
	}

	if found, path := checkSystemPath(name, exts); found {
		return true, path
	}

	return false, ""
}

func checkPath(name, dir string, exts []string) (bool, string) {
	fullPath := filepath.Join(dir, name)
	if exts == nil {
		if _, err := os.Stat(fullPath); err == nil {
			return true, fullPath
		}
	}

	for _, ext := range exts {
		fullPath := filepath.Join(dir, name+ext)
		if _, err := os.Stat(fullPath); err == nil {
			return true, fullPath
		}
	}

	return false, ""
}

func checkSystemPath(name string, exts []string) (bool, string) {
	for _, ext := range exts {
		path, err := exec.LookPath(name + ext)
		if err == nil {
			return true, path
		}
	}

	path, err := exec.LookPath(name)
	if err == nil {
		return true, path
	}

	return false, ""
}

func getExtensions() []string {
	if platform == PlatformWindows {
		return []string{".exe", ".bat", ".cmd", ".ps1"}
	}
	return nil
}

func NormalizeCommandName(name string) string {
	if platform == PlatformWindows && !strings.HasSuffix(name, ".exe") {
		return name + ".exe"
	}
	return name
}

func CommandExists(name string) bool {
	found, _ := FindCommand(name)
	return found
}

func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func MakeExecutable(path string) error {
	if platform != PlatformWindows {
		return os.Chmod(path, 0755)
	}
	return nil
}

type Dependency struct {
	Name        string
	Required    bool
	DisplayName string
	URL         string
	Extract     bool
}

func (d *Dependency) GetDownloadURL() string {
	if d.URL != "" {
		return d.URL
	}

	switch d.Name {
	case "yt-dlp":
		return getYtDlpURL()
	case "ffmpeg":
		return getFFmpegURL()
	case "deno":
		return getDenoURL()
	}
	return ""
}

func getYtDlpURL() string {
	switch platform {
	case PlatformWindows:
		return "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp.exe"
	case PlatformDarwin:
		return "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_macos"
	default:
		return "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp"
	}
}

func getFFmpegURL() string {
	switch platform {
	case PlatformWindows:
		return "https://github.com/BtbN/FFmpeg-Builds/releases/latest/download/ffmpeg-master-latest-win64-gpl-shared.zip"
	case PlatformDarwin:
		return "https://evermeet.codingfork.com/ffmpeg/ffmpeg"
	default:
		return "https://github.com/BtbN/FFmpeg-Builds/releases/latest/download/ffmpeg-master-latest-linux64-gpl-shared.tar.xz"
	}
}

func getDenoURL() string {
	switch platform {
	case PlatformWindows:
		return "https://deno.land/install.ps1"
	default:
		return "https://deno.land/install.sh"
	}
}

func GetInstallMessage(name string) string {
	switch name {
	case "yt-dlp":
		return fmt.Sprintf("Install yt-dlp:\n  Linux/macOS: pip install yt-dlp\n  Windows: pip install yt-dlp\n  Or: https://github.com/yt-dlp/yt-dlp/releases")
	case "ffmpeg":
		return fmt.Sprintf("Install ffmpeg:\n  Linux: sudo apt install ffmpeg\n  macOS: brew install ffmpeg\n  Windows: https://ffmpeg.org/download.html")
	case "deno":
		return fmt.Sprintf("Install deno:\n  Linux/macOS: curl -fsSL https://deno.land/install.sh | sh\n  Windows: irm https://deno.land/install.ps1 | iex\n  Or: https://deno.land/")
	}
	return fmt.Sprintf("Install %s", name)
}

func GetLocalPath(name string) string {
	localSrc := GetSrcDir()
	normalized := NormalizeCommandName(name)
	return filepath.Join(localSrc, normalized)
}
