package home

import (
	"errors"
	"os"
	"path/filepath"
)

type Paths struct {
	Home      string
	Toolchains string
	Metadata  string
	Config    string
}

func Resolve() (Paths, error) {
	home := os.Getenv("XGOUP_HOME")
	if home == "" {
		up := os.Getenv("USERPROFILE")
		if up != "" {
			home = filepath.Join(up, ".xgoup")
		} else {
			home = filepath.Join(os.TempDir(), ".xgoup")
		}
	}
	if !filepath.IsAbs(home) {
		abs, err := filepath.Abs(home)
		if err != nil {
			return Paths{}, err
		}
		home = abs
	}

	return Paths{
		Home:       home,
		Toolchains: filepath.Join(home, "toolchains"),
		Metadata:   filepath.Join(home, "metadata"),
		Config:     filepath.Join(home, "config.toml"),
	}, nil
}

func EnsureDirs(p Paths) error {
	if p.Home == "" {
		return errors.New("home is empty")
	}
	if err := os.MkdirAll(p.Toolchains, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(p.Metadata, 0o755); err != nil {
		return err
	}
	return nil
}

