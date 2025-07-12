package appdir

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const DefaultConfigSubdir = "kitops"
const PromptkitSubdir = ".promptkit"

// KitopsHome returns the base KITOPS_HOME directory based on the OS.
func KitopsHome() (string, error) {
	if v := os.Getenv("KITOPS_HOME"); v != "" {
		return v, nil
	}
	switch runtime.GOOS {
	case "linux":
		datahome := os.Getenv("XDG_DATA_HOME")
		if datahome == "" {
			userhome := os.Getenv("HOME")
			if userhome == "" {
				return "", fmt.Errorf("could not get $HOME directory")
			}
			datahome = filepath.Join(userhome, ".local", "share")
		}
		return filepath.Join(datahome, DefaultConfigSubdir), nil
	case "darwin":
		cacheDir, err := os.UserCacheDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(cacheDir, DefaultConfigSubdir), nil
	case "windows":
		appdata, err := os.UserCacheDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(appdata, DefaultConfigSubdir), nil
	default:
		return "", fmt.Errorf("unrecognized operating system")
	}
}

// PromptkitDir returns the path to the .promptkit directory under KITOPS_HOME.
func PromptkitDir() (string, error) {
	home, err := KitopsHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, PromptkitSubdir), nil
}

// SessionsDir returns the sessions directory under KITOPS_HOME.
func SessionsDir() (string, error) {
	dir, err := PromptkitDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "sessions"), nil
}

// SessionLogPath returns the path to today's session log file, creating the
// sessions directory if needed.
func SessionLogPath() (string, error) {
	sessionsDir, err := SessionsDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(sessionsDir, 0o755); err != nil {
		return "", err
	}
	name := "chat-" + time.Now().Format("2006-01-02") + ".jsonl"
	return filepath.Join(sessionsDir, name), nil
}
