package toolchain

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/fanfeilong/xgoup/internal/config"
	"github.com/fanfeilong/xgoup/internal/home"
	"github.com/fanfeilong/xgoup/internal/meta"
)

const DefaultSourceRepo = "https://github.com/goplus/xgo.git"

type InstallOptions struct {
	Method string
	Repo   string
	Ref    string
	Path   string
	Force  bool
}

func Init() (home.Paths, config.Config, error) {
	p, err := home.Resolve()
	if err != nil {
		return home.Paths{}, config.Config{}, err
	}
	if err := home.EnsureDirs(p); err != nil {
		return home.Paths{}, config.Config{}, err
	}

	cfg, err := readOrCreateConfig(p)
	if err != nil {
		return home.Paths{}, config.Config{}, err
	}
	return p, cfg, nil
}

func readOrCreateConfig(p home.Paths) (config.Config, error) {
	_, err := os.Stat(p.Config)
	if err == nil {
		c, err := config.Load(p.Config)
		if err != nil {
			return config.Config{}, err
		}
		if err := config.Validate(c); err != nil {
			return config.Config{}, err
		}
		return c, nil
	}
	if !os.IsNotExist(err) {
		return config.Config{}, err
	}

	c := config.Default(DefaultSourceRepo)

	// Seed from existing metadata if present.
	names, _ := meta.Names(p.Metadata)
	for _, n := range names {
		m, err := meta.Load(p.Metadata, n)
		if err != nil {
			continue
		}
		managed := strings.EqualFold(m.Managed, "true")
		c.Toolchains[n] = config.Toolchain{
			Kind:    m.Kind,
			Path:    m.RootPath,
			Managed: managed,
			Version: m.Version,
			Repo:    m.Repo,
			Ref:     m.Ref,
			Commit:  m.Commit,
		}
	}
	if c.DefaultToolchain == "" && len(c.Toolchains) > 0 {
		// Pick first stable ordering.
		var keys []string
		for k := range c.Toolchains {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		c.DefaultToolchain = keys[0]
	}
	if err := config.Save(p.Config, c); err != nil {
		return config.Config{}, err
	}
	return c, nil
}

func SaveConfig(p home.Paths, c config.Config) error {
	if err := config.Validate(c); err != nil {
		return err
	}
	return config.Save(p.Config, c)
}

func SetDefault(p home.Paths, c config.Config, name string) (config.Config, error) {
	if _, ok := c.Toolchains[name]; !ok {
		return config.Config{}, fmt.Errorf("toolchain not found: %s", name)
	}
	c.DefaultToolchain = name
	if err := SaveConfig(p, c); err != nil {
		return config.Config{}, err
	}
	return c, nil
}

func List(p home.Paths, c config.Config) ([]meta.ToolchainMeta, error) {
	names, err := meta.Names(p.Metadata)
	if err != nil {
		return nil, err
	}
	out := make([]meta.ToolchainMeta, 0, len(names))
	for _, n := range names {
		m, err := meta.Load(p.Metadata, n)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, nil
}

func Remove(p home.Paths, c config.Config, name string, purge bool) (config.Config, error) {
	m, err := meta.Load(p.Metadata, name)
	if err != nil {
		return config.Config{}, err
	}
	if purge && strings.EqualFold(m.Managed, "true") && m.RootPath != "" {
		_ = os.RemoveAll(m.RootPath)
	}
	_ = os.Remove(meta.FilePath(p.Metadata, name))

	delete(c.Toolchains, name)
	if c.DefaultToolchain == name {
		c.DefaultToolchain = ""
	}
	if c.DefaultToolchain == "" && len(c.Toolchains) > 0 {
		var keys []string
		for k := range c.Toolchains {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		c.DefaultToolchain = keys[0]
	}
	if err := SaveConfig(p, c); err != nil {
		return config.Config{}, err
	}
	return c, nil
}

func Install(p home.Paths, c config.Config, name string, opt InstallOptions) (config.Config, error) {
	if name == "" {
		return config.Config{}, errors.New("missing toolchain name")
	}
	if opt.Method == "" {
		opt.Method = "standard"
	}
	switch opt.Method {
	case "standard":
		return installStandard(p, c, name, opt)
	case "linked":
		return installLinked(p, c, name, opt)
	case "source":
		return installSource(p, c, name, opt)
	default:
		return config.Config{}, fmt.Errorf("unknown method: %s", opt.Method)
	}
}

func Update(p home.Paths, c config.Config, name string) (config.Config, error) {
	if name == "" {
		name = c.DefaultToolchain
	}
	if name == "" {
		return config.Config{}, errors.New("no toolchain specified and no default configured")
	}
	tc, ok := c.Toolchains[name]
	if !ok {
		return config.Config{}, fmt.Errorf("toolchain not found: %s", name)
	}
	switch tc.Kind {
	case "standard":
		return installStandard(p, c, name, InstallOptions{Method: "standard", Force: true})
	case "source":
		return installSourceUpdate(p, c, name)
	case "linked":
		return c, fmt.Errorf("linked toolchain update is a no-op: %s", name)
	default:
		return config.Config{}, fmt.Errorf("unknown toolchain kind: %s", tc.Kind)
	}
}

func Doctor(p home.Paths, c config.Config) error {
	// Basic checks.
	if err := config.Validate(c); err != nil {
		return err
	}
	// Ensure metadata entries are loadable and xgo exists.
	names, _ := meta.Names(p.Metadata)
	for _, n := range names {
		m, err := meta.Load(p.Metadata, n)
		if err != nil {
			return fmt.Errorf("metadata load failed for %s: %w", n, err)
		}
		if m.XgoBin != "" {
			if _, err := os.Stat(m.XgoBin); err != nil {
				return fmt.Errorf("xgo missing for %s: %s", n, m.XgoBin)
			}
		}
	}
	return nil
}

// --- install implementations ---

func installLinked(p home.Paths, c config.Config, name string, opt InstallOptions) (config.Config, error) {
	if opt.Path == "" {
		return config.Config{}, errors.New("linked method requires --path")
	}
	abs, err := filepath.Abs(opt.Path)
	if err != nil {
		return config.Config{}, err
	}
	xgo := filepath.Join(abs, "bin", "xgo.exe")
	if _, err := os.Stat(xgo); err != nil {
		return config.Config{}, fmt.Errorf("linked path missing xgo: %s", xgo)
	}
	if _, err := os.Stat(meta.FilePath(p.Metadata, name)); err == nil && !opt.Force {
		return config.Config{}, fmt.Errorf("toolchain already exists: %s (use --force)", name)
	}
	m := meta.ToolchainMeta{
		Name:         name,
		Kind:         "linked",
		RootPath:     abs,
		XgoBin:       xgo,
		XgoRootValue: abs,
		Managed:      "false",
		Version:      detectVersionFromXgo(xgo, abs),
	}
	if err := meta.Save(p.Metadata, name, m); err != nil {
		return config.Config{}, err
	}
	c.Toolchains[name] = config.Toolchain{Kind: "linked", Path: abs, Managed: false, Version: m.Version}
	if c.DefaultToolchain == "" {
		c.DefaultToolchain = name
	}
	if err := SaveConfig(p, c); err != nil {
		return config.Config{}, err
	}
	return c, nil
}

func installStandard(p home.Paths, c config.Config, name string, opt InstallOptions) (config.Config, error) {
	if _, err := os.Stat(meta.FilePath(p.Metadata, name)); err == nil && !opt.Force {
		return config.Config{}, fmt.Errorf("toolchain already exists: %s (use --force)", name)
	}

	arch, err := windowsArch()
	if err != nil {
		return config.Config{}, err
	}

	tag, asset, err := latestXgoWindowsAsset(arch)
	if err != nil {
		return config.Config{}, err
	}
	url := fmt.Sprintf("https://github.com/goplus/xgo/releases/download/%s/%s", tag, asset)

	tmpZip := filepath.Join(os.TempDir(), fmt.Sprintf("xgo-%d.zip", time.Now().UnixNano()))
	defer os.Remove(tmpZip)
	if err := httpGetToFile(url, tmpZip); err != nil {
		return config.Config{}, err
	}

	target := filepath.Join(p.Toolchains, name)
	if opt.Force {
		_ = os.RemoveAll(target)
	}
	if err := unzipAll(tmpZip, target); err != nil {
		return config.Config{}, err
	}

	xgoExe, xgoroot, err := findXgoAndRoot(target)
	if err != nil {
		return config.Config{}, err
	}

	m := meta.ToolchainMeta{
		Name:         name,
		Kind:         "standard",
		RootPath:     xgoroot,
		XgoBin:       xgoExe,
		XgoRootValue: xgoroot,
		Managed:      "true",
		Version:      detectVersionFromXgo(xgoExe, xgoroot),
	}
	if m.Version == "" {
		m.Version = tag
	}

	if err := meta.Save(p.Metadata, name, m); err != nil {
		return config.Config{}, err
	}
	c.Toolchains[name] = config.Toolchain{Kind: "standard", Path: xgoroot, Managed: true, Version: m.Version}
	if c.DefaultToolchain == "" {
		c.DefaultToolchain = name
	}
	if err := SaveConfig(p, c); err != nil {
		return config.Config{}, err
	}
	return c, nil
}

func installSource(p home.Paths, c config.Config, name string, opt InstallOptions) (config.Config, error) {
	repo := opt.Repo
	if repo == "" {
		repo = DefaultSourceRepo
	}
	if _, err := os.Stat(meta.FilePath(p.Metadata, name)); err == nil && !opt.Force {
		return config.Config{}, fmt.Errorf("toolchain already exists: %s (use --force)", name)
	}
	if _, err := exec.LookPath("git"); err != nil {
		return config.Config{}, errors.New("missing required command: git")
	}
	bash, err := findBash()
	if err != nil {
		return config.Config{}, err
	}

	target := filepath.Join(p.Toolchains, name)
	if opt.Force {
		_ = os.RemoveAll(target)
	}

	if err := execCmd("", "git", "clone", repo, target); err != nil {
		return config.Config{}, err
	}
	if opt.Ref != "" {
		if err := execCmd("", "git", "-C", target, "checkout", opt.Ref); err != nil {
			return config.Config{}, err
		}
	}

	// Build via bash -lc ./all.bash
	if err := execCmd(target, bash, "-lc", "./all.bash"); err != nil {
		return config.Config{}, err
	}

	xgoExe, xgoroot, err := findXgoAndRoot(target)
	if err != nil {
		return config.Config{}, err
	}

	commit := ""
	if out, err := exec.Command("git", "-C", target, "rev-parse", "HEAD").CombinedOutput(); err == nil {
		commit = strings.TrimSpace(string(out))
	}

	m := meta.ToolchainMeta{
		Name:         name,
		Kind:         "source",
		RootPath:     xgoroot,
		XgoBin:       xgoExe,
		XgoRootValue: xgoroot,
		Managed:      "true",
		Repo:         repo,
		Ref:          opt.Ref,
		Commit:       commit,
		Version:      detectVersionFromXgo(xgoExe, xgoroot),
	}
	if err := meta.Save(p.Metadata, name, m); err != nil {
		return config.Config{}, err
	}
	c.Toolchains[name] = config.Toolchain{Kind: "source", Path: xgoroot, Managed: true, Version: m.Version, Repo: repo, Ref: opt.Ref, Commit: commit}
	if c.DefaultToolchain == "" {
		c.DefaultToolchain = name
	}
	if err := SaveConfig(p, c); err != nil {
		return config.Config{}, err
	}
	return c, nil
}

func installSourceUpdate(p home.Paths, c config.Config, name string) (config.Config, error) {
	tc := c.Toolchains[name]
	if tc.Path == "" {
		return config.Config{}, errors.New("source toolchain missing path")
	}
	if _, err := exec.LookPath("git"); err != nil {
		return config.Config{}, errors.New("missing required command: git")
	}
	bash, err := findBash()
	if err != nil {
		return config.Config{}, err
	}
	// best-effort fetch + pull + rebuild
	_ = execCmd("", "git", "-C", tc.Path, "fetch", "--tags", "--prune")
	_ = execCmd("", "git", "-C", tc.Path, "pull", "--ff-only")
	if err := execCmd(tc.Path, bash, "-lc", "./all.bash"); err != nil {
		return config.Config{}, err
	}
	xgoExe, xgoroot, err := findXgoAndRoot(tc.Path)
	if err != nil {
		return config.Config{}, err
	}
	m, _ := meta.Load(p.Metadata, name)
	m.RootPath = xgoroot
	m.XgoBin = xgoExe
	m.XgoRootValue = xgoroot
	m.Version = detectVersionFromXgo(xgoExe, xgoroot)
	if out, err := exec.Command("git", "-C", tc.Path, "rev-parse", "HEAD").CombinedOutput(); err == nil {
		m.Commit = strings.TrimSpace(string(out))
	}
	if err := meta.Save(p.Metadata, name, m); err != nil {
		return config.Config{}, err
	}
	tc.Path = xgoroot
	tc.Version = m.Version
	tc.Commit = m.Commit
	c.Toolchains[name] = tc
	if err := SaveConfig(p, c); err != nil {
		return config.Config{}, err
	}
	return c, nil
}

// --- helpers ---

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

func latestXgoWindowsAsset(arch string) (tag string, asset string, err error) {
	req, err := http.NewRequest("GET", "https://api.github.com/repos/goplus/xgo/releases/latest", nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "xgoup")
	if tok := os.Getenv("GITHUB_TOKEN"); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", "", fmt.Errorf("GET %s: status=%s body=%q", req.URL.String(), resp.Status, string(b))
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", "", err
	}
	var v struct {
		Tag string `json:"tag_name"`
	}
	if err := json.Unmarshal(b, &v); err != nil {
		return "", "", err
	}
	if v.Tag == "" {
		return "", "", errors.New("failed to resolve xgo latest release tag")
	}
	ver := strings.TrimPrefix(v.Tag, "v")
	return v.Tag, fmt.Sprintf("xgo%s.windows-%s.zip", ver, arch), nil
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

func unzipAll(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}
	for _, f := range r.File {
		p := filepath.Join(destDir, filepath.FromSlash(f.Name))
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(p, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.Create(p)
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
	}
	return nil
}

func findXgoAndRoot(root string) (xgoExe string, xgoRoot string, err error) {
	// locate xgo.exe
	var found string
	err = filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.EqualFold(d.Name(), "xgo.exe") {
			found = p
			return errFound // sentinel to stop
		}
		return nil
	})
	if err != nil && !errors.Is(err, errFound) {
		return "", "", err
	}
	if found == "" {
		return "", "", errors.New("xgo.exe not found after extraction/build")
	}
	binDir := filepath.Dir(found)
	cand := binDir
	if strings.EqualFold(filepath.Base(binDir), "bin") {
		cand = filepath.Dir(binDir)
	}
	// Release zips have pkg/ or src/; a from-source build may only have all.bash at repo root.
	isValid := func(d string) bool {
		if d == "" {
			return false
		}
		if _, err := os.Stat(filepath.Join(d, "bin", "xgo.exe")); err != nil {
			return false
		}
		if _, err := os.Stat(filepath.Join(d, "pkg")); err == nil {
			return true
		}
		if _, err := os.Stat(filepath.Join(d, "src")); err == nil {
			return true
		}
		if _, err := os.Stat(filepath.Join(d, "all.bash")); err == nil {
			return true
		}
		return false
	}
	for !isValid(cand) {
		parent := filepath.Dir(cand)
		if parent == cand {
			break
		}
		cand = parent
	}
	if !isValid(cand) {
		return found, cand, fmt.Errorf("XGOROOT (%s) is not valid", cand)
	}
	return found, cand, nil
}

var errFound = errors.New("found")

func detectVersionFromXgo(xgoExe, xgoRoot string) string {
	// xgo prints version on `version`; ignore errors.
	// Clear inherited XGOROOT (e.g. user set XGOROOT=C:\) so xgo can resolve its own root.
	cmd := exec.Command(xgoExe, "version")
	cmd.Env = envForXgoCLI(xgoRoot)
	out, _ := cmd.CombinedOutput()
	s := strings.TrimSpace(string(out))
	if s == "" {
		return ""
	}
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " ")
}

// envForXgoCLI returns process env without a bogus XGOROOT, then sets XGOROOT when known.
func envForXgoCLI(xgoRoot string) []string {
	out := stripXGOROOTFromEnviron()
	if xgoRoot != "" {
		out = append(out, "XGOROOT="+xgoRoot)
	}
	return out
}

func stripXGOROOTFromEnviron() []string {
	var out []string
	for _, e := range os.Environ() {
		if strings.HasPrefix(strings.ToUpper(e), "XGOROOT=") {
			continue
		}
		out = append(out, e)
	}
	return out
}

func findBash() (string, error) {
	if p, err := exec.LookPath("bash.exe"); err == nil {
		return p, nil
	}
	if p, err := exec.LookPath("bash"); err == nil {
		return p, nil
	}
	// Common Git-Bash location
	common := []string{
		filepath.Join(os.Getenv("ProgramFiles"), "Git", "bin", "bash.exe"),
		filepath.Join(os.Getenv("ProgramFiles(x86)"), "Git", "bin", "bash.exe"),
	}
	for _, p := range common {
		if p == "" {
			continue
		}
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", errors.New("bash.exe not found; install Git for Windows (Git-Bash) or MSYS2 and ensure bash is on PATH")
}

func execCmd(dir, exe string, args ...string) error {
	cmd := exec.Command(exe, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Env = stripXGOROOTFromEnviron()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

