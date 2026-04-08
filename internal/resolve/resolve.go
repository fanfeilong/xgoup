package resolve

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fanfeilong/xgoup/internal/config"
	"github.com/fanfeilong/xgoup/internal/home"
	"github.com/fanfeilong/xgoup/internal/meta"
	"github.com/pelletier/go-toml/v2"
)

type Result struct {
	Name    string
	Meta    meta.ToolchainMeta
	XgoExe  string
	XgoRoot string
	BinDir  string
}

func ResolveToolchain(p home.Paths, c config.Config, cliName string, cwd string) (Result, error) {
	if cliName != "" {
		return byName(p, c, cliName)
	}
	if env := os.Getenv("XGO_TOOLCHAIN"); env != "" {
		return byName(p, c, env)
	}
	if cwd != "" {
		if name, ok := projectOverride(cwd); ok {
			return byName(p, c, name)
		}
	}
	if c.DefaultToolchain != "" {
		return byName(p, c, c.DefaultToolchain)
	}
	return Result{}, errors.New("no toolchain selected (set default or use --toolchain)")
}

func projectOverride(start string) (string, bool) {
	dir := start
	for {
		p := filepath.Join(dir, "xgo-toolchain.toml")
		b, err := os.ReadFile(p)
		if err == nil {
			var v struct {
				Toolchain string `toml:"toolchain"`
			}
			if err := toml.Unmarshal(b, &v); err == nil && strings.TrimSpace(v.Toolchain) != "" {
				return strings.TrimSpace(v.Toolchain), true
			}
			return "", false
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

func byName(p home.Paths, c config.Config, name string) (Result, error) {
	if _, ok := c.Toolchains[name]; !ok {
		return Result{}, fmt.Errorf("toolchain not found: %s", name)
	}
	m, err := meta.Load(p.Metadata, name)
	if err != nil {
		return Result{}, err
	}
	if m.XgoBin == "" {
		// fallback: assume <root>/bin/xgo.exe
		if m.RootPath != "" {
			m.XgoBin = filepath.Join(m.RootPath, "bin", "xgo.exe")
		}
	}
	if m.XgoRootValue == "" {
		m.XgoRootValue = m.RootPath
	}
	return Result{
		Name:    name,
		Meta:    m,
		XgoExe:  m.XgoBin,
		XgoRoot: m.XgoRootValue,
		BinDir:  filepath.Dir(m.XgoBin),
	}, nil
}

