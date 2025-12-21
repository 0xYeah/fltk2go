//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/0xYeah/fltk2go/config"
)

type fltk2goManifest struct {
	Module      string `json:"module"`
	FLTKVersion string `json:"fltk_version"`
	Target      struct {
		GOOS    string `json:"goos"`
		OutArch string `json:"out_arch"`
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

func listLibFilenames(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Printf("Error reading dir %s, %v\n", dir, err)
		os.Exit(1)
	}

	var libs []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasPrefix(name, "libfltk") && strings.HasSuffix(name, ".a") {
			libs = append(libs, name)
		}
	}
	sort.Strings(libs)
	return libs
}

func fileExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && st.Mode().IsRegular()
}

func getFltkGitRev(ctx *buildCtx) string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = ctx.fltkSource
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func inferToolchain(goos string) string {
	switch goos {
	case "windows":
		return "cmake+mingw"
	case "darwin":
		return "cmake+clang"
	default:
		return "cmake"
	}
}

// writeManifestForTarget writes manifest to: libs/fltk/<goos>/<outArch>/fltk2go.manifest.json
func writeManifestForTarget(ctx *buildCtx, outGOOS, outArch string) {
	targetDir := filepath.Join("libs", "fltk", outGOOS, outArch)
	if err := os.MkdirAll(targetDir, 0750); err != nil {
		fmt.Printf("Error creating dir %s, %v\n", targetDir, err)
		os.Exit(1)
	}

	libs := listLibFilenames(targetDir)
	hasCfg := fileExists(filepath.Join(targetDir, "FL", "fl_config.h"))

	var m fltk2goManifest
	// 如果你 module path 不是这个，改成你 go.mod 里的 module 行
	m.Module = "github.com/0xYeah/fltk2go"
	m.FLTKVersion = config.FLTKPreBuildVersion
	m.Target.GOOS = outGOOS
	m.Target.OutArch = outArch

	m.Build.Toolchain = inferToolchain(outGOOS)
	m.Build.Date = time.Now().UTC().Format(time.RFC3339)
	m.Build.GitRev = getFltkGitRev(ctx)

	m.Artifacts.Libs = libs
	m.Artifacts.HasFlConfig = hasCfg

	b, err := json.MarshalIndent(&m, "", "  ")
	if err != nil {
		fmt.Printf("Error marshal manifest, %v\n", err)
		os.Exit(1)
	}

	dst := filepath.Join(targetDir, "fltk2go.manifest.json")
	if err := os.WriteFile(dst, b, 0644); err != nil {
		fmt.Printf("Error writing %s, %v\n", dst, err)
		os.Exit(1)
	}

	fmt.Printf("Wrote manifest: %s\n", dst)
}
