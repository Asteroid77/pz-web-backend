package pzpaths

import (
	"os"
	"path/filepath"
)

type BaseDirs struct {
	DataDir string
	GameDir string
}

type ResolveBaseDirsOptions struct {
	DevMode bool
	CWD     string

	EnvDataDir    string
	EnvInstallDir string
}

func ResolveBaseDirs(opts ResolveBaseDirsOptions) BaseDirs {
	if opts.DevMode {
		root := opts.CWD
		if root == "" {
			root = "."
		}
		return BaseDirs{
			DataDir: firstExisting(
				filepath.Join(root, "testdata/mock_zomboid_full/mock_zomboid"),
				filepath.Join(root, "testdata/mock_zomboid"),
			),
			GameDir: firstExisting(
				filepath.Join(root, "testdata/mock_media_full/mock_media"),
				filepath.Join(root, "testdata/mock_media"),
			),
		}
	}

	dataDir := opts.EnvDataDir
	if dataDir == "" {
		dataDir = "/home/steam/Zomboid"
	}

	installDir := opts.EnvInstallDir
	if installDir == "" {
		installDir = "/opt/pzserver"
	}

	return BaseDirs{
		DataDir: dataDir,
		GameDir: filepath.Join(installDir, "media"),
	}
}

func firstExisting(paths ...string) string {
	for _, p := range paths {
		if p == "" {
			continue
		}
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	if len(paths) == 0 {
		return ""
	}
	return paths[0]
}
