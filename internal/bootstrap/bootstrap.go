// Package bootstrap implements `xgoup self install` (formerly the separate xgoup-init binary).
package bootstrap

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Options configures a self-install from release assets or a local archive.
type Options struct {
	Repo       string // owner/name, default fanfeilong/xgoup
	Version    string // "latest" or "v0.1.0"
	ZipPath    string // local archive: Windows .zip, Unix .tar.gz (optional)
	ModifyPath bool   // best-effort add ~/.xgoup/bin to user PATH
	Logf       func(format string, args ...any)
}

func logf(o Options, format string, args ...any) {
	if o.Logf != nil {
		o.Logf(format, args...)
	}
}

// Install downloads (or uses --zip) the release payload for this OS/arch and
// installs xgoup into ~/.xgoup/bin, then optionally updates user PATH.
func Install(o Options) error {
	if o.Repo == "" {
		o.Repo = "fanfeilong/xgoup"
	}
	if o.Version == "" {
		o.Version = "latest"
	}
	home := defaultHome()
	binDir := filepath.Join(home, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return err
	}

	switch runtime.GOOS {
	case "windows":
		return installWindows(binDir, o)
	case "darwin", "linux":
		return installUnix(binDir, o)
	default:
		return fmt.Errorf("self install: unsupported GOOS %s", runtime.GOOS)
	}
}

func defaultHome() string {
	if h := os.Getenv("XGOUP_HOME"); h != "" {
		return h
	}
	if runtime.GOOS == "windows" {
		up := os.Getenv("USERPROFILE")
		if up == "" {
			return filepath.Join(os.TempDir(), ".xgoup")
		}
		return filepath.Join(up, ".xgoup")
	}
	h, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), ".xgoup")
	}
	return filepath.Join(h, ".xgoup")
}

func installWindows(binDir string, o Options) error {
	arch, err := detectArch()
	if err != nil {
		return err
	}
	tryURLs := buildTryURLs(o.Repo, o.Version, arch, "windows")

	tmp := filepath.Join(os.TempDir(), fmt.Sprintf("xgoup-%d.zip", time.Now().UnixNano()))
	var lastErr error
	zipPath := o.ZipPath
	if zipPath != "" {
		if _, err := os.Stat(zipPath); err != nil {
			return err
		}
		tmp = zipPath
	} else {
		defer func() { _ = os.Remove(tmp) }()
		for _, u := range tryURLs {
			logf(o, "[xgoup] download: %s", u)
			if err := httpGetToFile(u, tmp); err != nil {
				lastErr = err
				logf(o, "[xgoup] download failed: %v", err)
				continue
			}
			lastErr = nil
			break
		}
		if lastErr != nil {
			return lastErr
		}
	}

	want := map[string]bool{"xgoup.exe": true}
	if err := unzipSelected(tmp, binDir, want); err != nil {
		return fmt.Errorf("unzip: %w", err)
	}
	logf(o, "[xgoup] installed: %s", binDir)

	for _, legacy := range []string{"xgoup.ps1", "xgoup.cmd"} {
		p := filepath.Join(binDir, legacy)
		if err := os.Remove(p); err == nil {
			logf(o, "[xgoup] removed legacy: %s", p)
		}
	}

	if o.ModifyPath {
		if err := addToUserPathWindows(binDir); err != nil {
			logf(o, "[xgoup] PATH update skipped/failed: %v", err)
		} else {
			logf(o, "[xgoup] updated user PATH")
		}
	}

	cmdPath := filepath.Join(binDir, "xgoup.exe")
	logf(o, "[xgoup] try: %s --version", cmdPath)
	out, err := exec.Command(cmdPath, "--version").CombinedOutput()
	if text := strings.TrimSpace(string(out)); text != "" {
		fmt.Println(text)
	}
	if err != nil {
		logf(o, "[xgoup] xgoup exit: %v", err)
	}
	return nil
}

func installUnix(binDir string, o Options) error {
	arch, err := detectArch()
	if err != nil {
		return err
	}
	goos := runtime.GOOS
	tryURLs := buildTryURLsUnix(o.Repo, o.Version, goos, arch)

	tmp := filepath.Join(os.TempDir(), fmt.Sprintf("xgoup-%d.tgz", time.Now().UnixNano()))

	var lastErr error
	if o.ZipPath != "" {
		if _, err := os.Stat(o.ZipPath); err != nil {
			return err
		}
		tmp = o.ZipPath
	} else {
		defer func() { _ = os.Remove(tmp) }()
		for _, u := range tryURLs {
			logf(o, "[xgoup] download: %s", u)
			if err := httpGetToFile(u, tmp); err != nil {
				lastErr = err
				logf(o, "[xgoup] download failed: %v", err)
				continue
			}
			lastErr = nil
			break
		}
		if lastErr != nil {
			return lastErr
		}
	}

	if err := extractTarGzXgoup(tmp, filepath.Join(binDir, "xgoup")); err != nil {
		return fmt.Errorf("extract: %w", err)
	}
	if err := os.Chmod(filepath.Join(binDir, "xgoup"), 0o755); err != nil {
		return err
	}
	logf(o, "[xgoup] installed: %s", binDir)

	if o.ModifyPath {
		if err := appendUnixUserPath(binDir); err != nil {
			logf(o, "[xgoup] PATH hint failed: %v", err)
		}
	}

	cmdPath := filepath.Join(binDir, "xgoup")
	logf(o, "[xgoup] try: %s --version", cmdPath)
	out, err := exec.Command(cmdPath, "--version").CombinedOutput()
	if text := strings.TrimSpace(string(out)); text != "" {
		fmt.Println(text)
	}
	if err != nil {
		logf(o, "[xgoup] xgoup exit: %v", err)
	}
	return nil
}

func buildTryURLs(repo, version, arch, goos string) []string {
	var try []string
	if version != "" && version != "latest" {
		try = append(try, fmt.Sprintf("https://github.com/%s/releases/download/%s/xgoup-%s-windows-%s.zip", repo, version, version, arch))
	}
	try = append(try, fmt.Sprintf("https://github.com/%s/releases/latest/download/xgoup-windows-%s.zip", repo, arch))
	if version == "latest" {
		if tag, err := getLatestReleaseTag(repo); err == nil {
			try = append(try, fmt.Sprintf("https://github.com/%s/releases/download/%s/xgoup-%s-windows-%s.zip", repo, tag, tag, arch))
		}
	}
	return try
}

func buildTryURLsUnix(repo, version, goos, arch string) []string {
	var try []string
	if version != "" && version != "latest" {
		try = append(try, fmt.Sprintf("https://github.com/%s/releases/download/%s/xgoup-%s-%s-%s.tar.gz", repo, version, version, goos, arch))
	}
	if version == "latest" {
		if tag, err := getLatestReleaseTag(repo); err == nil {
			try = append(try, fmt.Sprintf("https://github.com/%s/releases/download/%s/xgoup-%s-%s-%s.tar.gz", repo, tag, tag, goos, arch))
		}
	}
	return try
}

func detectArch() (string, error) {
	switch runtime.GOARCH {
	case "amd64":
		return "amd64", nil
	case "arm64":
		return "arm64", nil
	default:
		return "", fmt.Errorf("unsupported arch: %s", runtime.GOARCH)
	}
}

func getLatestReleaseTag(repo string) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "xgoup")
	if tok := os.Getenv("GITHUB_TOKEN"); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("GET %s: status=%s body=%q", req.URL.String(), resp.Status, string(b))
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}
	var v struct {
		Tag string `json:"tag_name"`
	}
	if err := json.Unmarshal(b, &v); err != nil {
		return "", err
	}
	if v.Tag == "" {
		return "", errors.New("empty tag_name")
	}
	return v.Tag, nil
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

func unzipSelected(zipPath, destDir string, want map[string]bool) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}
	found := map[string]bool{}
	for _, f := range r.File {
		name := filepath.Base(strings.ReplaceAll(f.Name, "\\", "/"))
		if !want[name] {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		outPath := filepath.Join(destDir, name)
		out, err := os.Create(outPath)
		if err != nil {
			rc.Close()
			return err
		}
		if _, err := io.Copy(out, rc); err != nil {
			out.Close()
			rc.Close()
			return err
		}
		out.Close()
		rc.Close()
		found[name] = true
	}
	for k := range want {
		if !found[k] {
			return fmt.Errorf("missing %q in zip", k)
		}
	}
	return nil
}

func extractTarGzXgoup(tgzPath, destBin string) error {
	f, err := os.Open(tgzPath)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		base := filepath.Base(hdr.Name)
		if base != "xgoup" {
			continue
		}
		out, err := os.Create(destBin)
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, tr); err != nil {
			out.Close()
			return err
		}
		return out.Close()
	}
	return errors.New("no file named xgoup in archive")
}

func addToUserPathWindows(binDir string) error {
	ps := fmt.Sprintf(`[Environment]::SetEnvironmentVariable("Path", (([Environment]::GetEnvironmentVariable("Path","User") + ";%s").Trim(";")), "User")`, strings.ReplaceAll(binDir, `"`, `""`))
	cmd := exec.Command("powershell.exe", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", ps)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("powershell PATH update: %w (%s)", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func appendUnixUserPath(binDir string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	rc := filepath.Join(home, ".profile")
	line := fmt.Sprintf(`export PATH="%s:$PATH"`, binDir)
	b, err := os.ReadFile(rc)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	s := string(b)
	if strings.Contains(s, binDir) {
		return nil
	}
	f, err := os.OpenFile(rc, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	if len(s) > 0 && !strings.HasSuffix(s, "\n") {
		if _, err := f.WriteString("\n"); err != nil {
			return err
		}
	}
	_, err = fmt.Fprintf(f, "\n# xgoup self install\n%s\n", line)
	return err
}
