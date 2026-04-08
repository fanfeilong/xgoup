package config

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/pelletier/go-toml/v2"
)

type Settings struct {
	AutoSelfUpdate bool `toml:"auto_self_update"`
	AutoPathHint   bool `toml:"auto_path_hint"`
	Telemetry      bool `toml:"telemetry"`
}

type Registries struct {
	SourceRepo     string `toml:"source_repo"`
	ReleaseBaseURL string `toml:"release_base_url"`
}

type Toolchain struct {
	Kind    string `toml:"kind"`
	Path    string `toml:"path"`
	Managed bool   `toml:"managed"`
	Version string `toml:"version,omitempty"`
	Repo    string `toml:"repo,omitempty"`
	Ref     string `toml:"ref,omitempty"`
	Commit  string `toml:"commit,omitempty"`
}

type Config struct {
	Version         int                  `toml:"version"`
	DefaultToolchain string              `toml:"default_toolchain"`
	LastUpdateCheck string               `toml:"last_update_check,omitempty"`
	Settings        Settings             `toml:"settings"`
	Registries      Registries           `toml:"registries"`
	Toolchains      map[string]Toolchain `toml:"toolchains"`
}

func Default(sourceRepo string) Config {
	return Config{
		Version:          1,
		DefaultToolchain: "",
		LastUpdateCheck:  time.Now().UTC().Format(time.RFC3339),
		Settings: Settings{
			AutoSelfUpdate: true,
			AutoPathHint:   true,
			Telemetry:      false,
		},
		Registries: Registries{
			SourceRepo:     sourceRepo,
			ReleaseBaseURL: "https://github.com/<org>/xgoup/releases/download",
		},
		Toolchains: map[string]Toolchain{},
	}
}

func Load(path string) (Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var c Config
	if err := toml.Unmarshal(b, &c); err != nil {
		return Config{}, err
	}
	if c.Version == 0 {
		c.Version = 1
	}
	if c.Toolchains == nil {
		c.Toolchains = map[string]Toolchain{}
	}
	return c, nil
}

func Save(path string, c Config) error {
	if c.Version == 0 {
		c.Version = 1
	}
	if c.LastUpdateCheck == "" {
		c.LastUpdateCheck = time.Now().UTC().Format(time.RFC3339)
	}
	if c.Toolchains == nil {
		c.Toolchains = map[string]Toolchain{}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := toml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func Validate(c Config) error {
	if c.Version != 1 {
		return errors.New("unsupported config version")
	}
	if c.DefaultToolchain != "" {
		if _, ok := c.Toolchains[c.DefaultToolchain]; !ok {
			return errors.New("default_toolchain does not exist in [toolchains]")
		}
	}
	seen := map[string]bool{}
	for name, tc := range c.Toolchains {
		if name == "" {
			return errors.New("toolchain name is empty")
		}
		if !filepath.IsAbs(tc.Path) {
			return errors.New("toolchain path must be absolute: " + name)
		}
		if seen[tc.Path] {
			return errors.New("duplicate toolchain path: " + tc.Path)
		}
		seen[tc.Path] = true
		if tc.Kind == "linked" && tc.Managed {
			return errors.New("linked toolchain must have managed=false: " + name)
		}
		if tc.Kind == "source" && tc.Repo == "" {
			return errors.New("source toolchain requires repo: " + name)
		}
	}
	return nil
}

