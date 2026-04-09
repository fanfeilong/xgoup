package meta

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type ToolchainMeta struct {
	Name        string
	Kind        string
	RootPath    string
	XgoBin      string
	XgoRootValue string
	Managed     string
	Repo        string
	Ref         string
	Commit      string
	Version     string
}

func FilePath(metaDir, name string) string {
	return filepath.Join(metaDir, name+".env")
}

func Names(metaDir string) ([]string, error) {
	ents, err := os.ReadDir(metaDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	var out []string
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		n := e.Name()
		if strings.HasSuffix(n, ".env") {
			out = append(out, strings.TrimSuffix(n, ".env"))
		}
	}
	sort.Strings(out)
	return out, nil
}

func Load(metaDir, name string) (ToolchainMeta, error) {
	p := FilePath(metaDir, name)
	f, err := os.Open(p)
	if err != nil {
		return ToolchainMeta{}, err
	}
	defer f.Close()

	m := ToolchainMeta{Name: name}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		v = parseSingleQuoted(v)
		switch k {
		case "NAME":
			m.Name = v
		case "KIND":
			m.Kind = v
		case "ROOT_PATH":
			m.RootPath = v
		case "XGO_BIN":
			m.XgoBin = v
		case "XGOROOT_VALUE":
			m.XgoRootValue = v
		case "MANAGED":
			m.Managed = v
		case "REPO":
			m.Repo = v
		case "REF":
			m.Ref = v
		case "COMMIT":
			m.Commit = v
		case "VERSION":
			m.Version = v
		}
	}
	if err := sc.Err(); err != nil {
		return ToolchainMeta{}, err
	}
	if m.Name == "" {
		m.Name = name
	}
	return m, nil
}

func Save(metaDir, name string, m ToolchainMeta) error {
	p := FilePath(metaDir, name)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer f.Close()

	// Match bash script's key ordering for diffs/debuggability.
	lines := []struct {
		k string
		v string
	}{
		{"NAME", or(m.Name, name)},
		{"KIND", m.Kind},
		{"ROOT_PATH", m.RootPath},
		{"XGO_BIN", m.XgoBin},
		{"XGOROOT_VALUE", m.XgoRootValue},
		{"MANAGED", m.Managed},
		{"REPO", m.Repo},
		{"REF", m.Ref},
		{"COMMIT", m.Commit},
		{"VERSION", m.Version},
	}
	for _, kv := range lines {
		if _, err := fmt.Fprintf(f, "%s='%s'\n", kv.k, escapeSingleQuotes(kv.v)); err != nil {
			return err
		}
	}
	return nil
}

func or(v, def string) string {
	if v != "" {
		return v
	}
	return def
}

func parseSingleQuoted(v string) string {
	// Expected: '...'
	if len(v) >= 2 && strings.HasPrefix(v, "'") && strings.HasSuffix(v, "'") {
		v = v[1 : len(v)-1]
	}
	// bash escape_sq used by script: s/'/'\\''/g -> literal sequence: '\'' in file
	v = strings.ReplaceAll(v, `'\''`, `'`)
	return v
}

func escapeSingleQuotes(s string) string {
	// Mirror bash escape_sq: ' -> '\'' (end quote, escape quote, reopen)
	return strings.ReplaceAll(s, `'`, `'\''`)
}

