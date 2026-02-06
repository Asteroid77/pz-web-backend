package configapp

import (
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"pz-web-backend/internal/config"
	"pz-web-backend/internal/infra/executil"
	"pz-web-backend/internal/infra/fs"
	"pz-web-backend/internal/infra/supervisor"
)

type Service struct {
	BaseDataDir string
	ServerName  string
	DevMode     bool

	Config config.Service
	FS     fs.FS
	Runner executil.Runner

	Restarter supervisor.Restarter
}

func (s Service) GetServerConfig(lang string) ([]config.Item, error) {
	serverName := s.resolvedServerName()
	path := filepath.Join(s.BaseDataDir, "Server", serverName+".ini")
	return s.Config.ParseServerINI(path, lang)
}

func (s Service) GetSandboxConfig(lang string) ([]config.Item, error) {
	serverName := s.resolvedServerName()
	path := filepath.Join(s.BaseDataDir, "Server", serverName+"_SandboxVars.lua")
	return s.Config.ParseSandboxLua(path, lang)
}

type SaveKind string

const (
	KindServer  SaveKind = "server"
	KindSandbox SaveKind = "sandbox"
)

func (s Service) Save(kind SaveKind, items []config.Item, restart bool) error {
	var path string
	var content string

	serverName := s.resolvedServerName()
	switch kind {
	case KindServer:
		path = filepath.Join(s.BaseDataDir, "Server", serverName+".ini")
		content = s.Config.GenerateServerINI(items)
	case KindSandbox:
		path = filepath.Join(s.BaseDataDir, "Server", serverName+"_SandboxVars.lua")
		content = s.Config.GenerateSandboxLua(items)
	default:
		return fmt.Errorf("invalid config kind: %s", kind)
	}

	if err := s.FS.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	if err := s.FS.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	if !s.DevMode && s.Runner != nil {
		_, _ = s.Runner.CombinedOutput("chown", "steam:steam", path)
	}

	if restart && s.Restarter != nil {
		return s.Restarter.RestartPZServer()
	}

	return nil
}

func (s Service) resolvedServerName() string {
	return ResolveServerName(s.FS, s.BaseDataDir, s.ServerName)
}

func ResolveServerName(fsys fs.FS, baseDataDir string, serverName string) string {
	name := strings.TrimSpace(serverName)
	if name != "" {
		name = path.Base(name)
		name = strings.TrimSuffix(name, ".ini")
		return name
	}

	serverDir := filepath.Join(baseDataDir, "Server")
	entries, err := fsys.ReadDir(serverDir)
	if err != nil {
		return "servertest"
	}

	var iniNames []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		n := e.Name()
		if !strings.HasSuffix(strings.ToLower(n), ".ini") {
			continue
		}
		iniNames = append(iniNames, n)
	}

	sort.Strings(iniNames)
	for _, n := range iniNames {
		if strings.EqualFold(n, "servertest.ini") {
			return "servertest"
		}
	}

	if len(iniNames) == 1 {
		return strings.TrimSuffix(iniNames[0], filepath.Ext(iniNames[0]))
	}

	return "servertest"
}
