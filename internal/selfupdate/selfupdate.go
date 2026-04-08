package selfupdate

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type Options struct {
	Repo string // owner/name
}

func Update(opt Options) (string, error) {
	repo := opt.Repo
	if repo == "" {
		repo = "fanfeilong/xgoup"
	}
	arch, err := windowsArch()
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("https://github.com/%s/releases/latest/download/xgoup-windows-%s.zip", repo, arch)

	tmpZip := filepath.Join(os.TempDir(), fmt.Sprintf("xgoup-self-%d.zip", time.Now().UnixNano()))
	defer os.Remove(tmpZip)
	if err := httpGetToFile(url, tmpZip); err != nil {
		return "", err
	}

	exe, err := os.Executable()
	if err != nil {
		return "", err
	}

	tmpExe := exe + ".new"
	if err := extractFileFromZip(tmpZip, "xgoup.exe", tmpExe); err != nil {
		// tolerate zips with nested paths; try any xgoup.exe
		if err2 := extractAnyBasenameFromZip(tmpZip, "xgoup.exe", tmpExe); err2 != nil {
			return "", err
		}
	}

	// Best-effort replace: on Windows, renaming over a running exe may fail.
	backup := exe + ".old"
	_ = os.Remove(backup)
	if err := os.Rename(exe, backup); err != nil {
		// Can't replace in-place; leave .new and instruct user.
		return tmpExe, fmt.Errorf("downloaded update to %s (failed to replace running exe: %v)", tmpExe, err)
	}
	if err := os.Rename(tmpExe, exe); err != nil {
		_ = os.Rename(backup, exe)
		return "", err
	}
	_ = os.Remove(backup)
	return exe, nil
}

func Uninstall(homeDir string, force bool) error {
	if homeDir == "" {
		return errors.New("homeDir is empty")
	}
	if !force {
		return errors.New("refusing to uninstall without --force")
	}
	return os.RemoveAll(homeDir)
}

func windowsArch() (string, error) {
	switch runtime.GOARCH {
	case "amd64":
		return "amd64", nil
	case "arm64":
		return "arm64", nil
	default:
		return "", fmt.Errorf("unsupported arch: %s", runtime.GOARCH)
	}
}

func httpGetToFile(url, path string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "xgoup")
	req.Header.Set("Accept", "application/octet-stream")
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("GET %s: status=%s body=%q", url, resp.Status, string(b))
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

func extractFileFromZip(zipPath, wantName, outPath string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		if filepath.Base(strings.ReplaceAll(f.Name, "\\", "/")) != wantName {
			continue
		}
		return writeZipFile(f, outPath)
	}
	return fmt.Errorf("missing %q in zip", wantName)
}

func extractAnyBasenameFromZip(zipPath, wantBase, outPath string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		if filepath.Base(strings.ReplaceAll(f.Name, "\\", "/")) == wantBase {
			return writeZipFile(f, outPath)
		}
	}
	return fmt.Errorf("missing %q in zip", wantBase)
}

func writeZipFile(f *zip.File, outPath string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	out, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, rc); err != nil {
		return err
	}
	return nil
}

