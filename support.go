package fltk2go

import (
	"embed"
	"encoding/json"
	"io/fs"
	"sort"
	"strings"
)

//go:embed libs/fltk/**/**/fltk2go.manifest.json
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
type SupportedLibrary struct {
	GOOS        string
	Arch        string
	FLTKVersion string
	Libraries   []string
}

// GetSupportedLibraries returns all FLTK static library bundles
// embedded in this fltk2go module.
func GetSupportedLibraries() ([]SupportedLibrary, error) {
	var result []SupportedLibrary

	err := fs.WalkDir(manifestsFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
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

		result = append(result, SupportedLibrary{
			GOOS:        m.Target.GOOS,
			Arch:        m.Target.OutArch,
			FLTKVersion: m.FLTKVersion,
			Libraries:   append([]string{}, m.Artifacts.Libs...), // defensive copy
		})
		return nil
	})

	if err != nil {
		return nil, err
	}

	// 稳定排序，方便输出 / 测试
	sort.Slice(result, func(i, j int) bool {
		if result[i].GOOS != result[j].GOOS {
			return result[i].GOOS < result[j].GOOS
		}
		return result[i].Arch < result[j].Arch
	})

	return result, nil
}
