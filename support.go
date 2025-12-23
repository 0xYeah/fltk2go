package fltk2go

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

//go:embed libs/*/*/*/*.manifest.json
var manifestsFS embed.FS

type Manifest struct {
	Module      string `json:"module"`
	FLTKVersion string `json:"fltk_version"`
	Target      struct {
		GOOS    string `json:"goos"`
		OutArch string `json:"out_arch"` // amd64 / arm64 / universal / ...
	} `json:"target"`
	Build struct {
		Toolchain string `json:"toolchain"`
		Date      string `json:"date"`
		GitRev    string `json:"git_rev"`
	} `json:"build"`
	Artifacts struct {
		Libs        []string `json:"libs"`
		HasFlConfig bool     `json:"has_fl_config"`
	} `json:"artifacts"`
}
type Library struct {
	GOOS        string
	Arch        string
	FLTKVersion string
	Libraries   []string
}

func extractLibNameFromManifestPath(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("empty path")
	}
	path = strings.ReplaceAll(path, "\\", "/")

	segs := make([]string, 0, 8)
	for _, p := range strings.Split(path, "/") {
		if p != "" {
			segs = append(segs, p)
		}
	}

	if len(segs) < 2 {
		return "", fmt.Errorf("invalid manifest path: %q", path)
	}
	if segs[0] != "libs" {
		return "", fmt.Errorf("unexpected manifest root %q (want libs): %q", segs[0], path)
	}

	return segs[1], nil
}

// GetSupportedLibraries returns all FLTK static library bundles
// embedded in this fltk2go module.
func GetSupportedLibraries() (map[string][]Library, error) {
	result := make(map[string][]Library)

	err := fs.WalkDir(manifestsFS, ".", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, "fltk2go.manifest.json") {
			return nil
		}

		data, err := fs.ReadFile(manifestsFS, path)
		if err != nil {
			return err
		}

		var m Manifest
		if err := json.Unmarshal(data, &m); err != nil {
			return err
		}

		libName, err := extractLibNameFromManifestPath(path)
		if err != nil {
			// 建议：这里直接返回错误，避免 silently 丢数据
			return err
		}

		item := Library{
			GOOS:        m.Target.GOOS,
			Arch:        m.Target.OutArch,
			FLTKVersion: m.FLTKVersion,
			Libraries:   append([]string(nil), m.Artifacts.Libs...), // defensive copy
		}

		result[libName] = append(result[libName], item)
		return nil
	})
	if err != nil {
		return nil, err
	}

	for name := range result {
		sort.Slice(result[name], func(i, j int) bool {
			a, b := result[name][i], result[name][j]
			if a.GOOS != b.GOOS {
				return a.GOOS < b.GOOS
			}
			return a.Arch < b.Arch
		})
	}

	return result, nil
}
