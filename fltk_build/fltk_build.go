//go:build ignore

package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/0xYeah/fltk2go/config"
)

var (
	FLTKBuildRoot      = getEnvOrDefault("FLTK_BUILD_ROOT", "build")
	OutputRoot         = getEnvOrDefault("FLTK_OUTPUT_ROOT", filepath.Join("build", "_install"))
	OutputArchOverride = os.Getenv("FLTK_OUTPUT_ARCH")
	CGOOutDir          = getEnvOrDefault("FLTK_CGO_OUT_DIR", "fltk_bridge")
	CGOPackage         = getEnvOrDefault("FLTK_CGO_PACKAGE", "fltk_bridge")
	FLTKPatchPath      = getEnvOrDefault("FLTK_PATCH_PATH", filepath.Join("fltk_build", "fltk-1.4.patch"))
	FinalRoot          = getEnvOrDefault("FLTK_FINAL_ROOT", filepath.Join("libs", "fltk"))
)

type buildCtx struct {
	goos       string
	goarch     string
	outArch    string
	outputRoot string
	libdir     string
	includeDir string
	finalRoot  string
	buildRoot  string
	fltkSource string
	cmakeBuild string
	currentDir string
	env        []string
}

func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func mustCheckEnv() {
	if runtime.GOOS == "" {
		fmt.Println("GOOS environment variable is empty")
		os.Exit(1)
	}
	if runtime.GOARCH == "" {
		fmt.Println("GOARCH environment variable is empty")
		os.Exit(1)
	}
	fmt.Printf("Building FLTK for OS: %s, architecture: %s\n", runtime.GOOS, runtime.GOARCH)
}

func mustCheckTool(tool string) {
	if _, err := exec.LookPath(tool); err != nil {
		fmt.Printf("Cannot find %s binary, %v\n", tool, err)
		os.Exit(1)
	}
}

func mustMkdirAll(path string, perm os.FileMode) {
	if err := os.MkdirAll(path, perm); err != nil {
		fmt.Printf("Could not create directory %s, %v\n", path, err)
		os.Exit(1)
	}
}

func mustGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Cannot get current directory, %v\n", err)
		os.Exit(1)
	}
	return wd
}

func newBuildCtx(goarch, outArch string) *buildCtx {
	goos := runtime.GOOS
	outputRoot := filepath.Clean(OutputRoot)
	buildRoot := filepath.Clean(FLTKBuildRoot)

	libdir := filepath.Join(outputRoot, "lib", goos, outArch)
	includeDir := filepath.Join(outputRoot, "include")
	finalRoot := filepath.Clean(FinalRoot)

	ctx := &buildCtx{
		goos:       goos,
		goarch:     goarch,
		outArch:    outArch,
		outputRoot: outputRoot,
		libdir:     libdir,
		includeDir: includeDir,
		finalRoot:  finalRoot,
		buildRoot:  buildRoot,
		fltkSource: filepath.Join(buildRoot, "fltk"),
		cmakeBuild: filepath.Join(buildRoot, fmt.Sprintf("fltk-cmake-%s-%s", goos, outArch)),
		currentDir: mustGetwd(),
	}

	if goos == "windows" {
		msys := os.Getenv("MSYS2_ROOT")
		if msys == "" {
			msys = `C:\msys64`
		}

		mingwBin := filepath.Join(msys, "mingw64", "bin")
		usrBin := filepath.Join(msys, "usr", "bin")

		mustFileExist(filepath.Join(mingwBin, "gcc.exe"))
		mustFileExist(filepath.Join(mingwBin, "g++.exe"))
		mustFileExist(filepath.Join(mingwBin, "mingw32-make.exe"))
		mustFileExist(filepath.Join(mingwBin, "windres.exe"))
		mustFileExist(filepath.Join(mingwBin, "cmake.exe"))

		oldPath := os.Getenv("PATH")
		// Windows PATH use ';'
		newPath := mingwBin + ";" + usrBin + ";" + oldPath

		ctx.env = append(os.Environ(),
			"PATH="+newPath,
			"MSYSTEM=MINGW64",
			"CHERE_INVOKING=1",
		)
	}

	return ctx
}

func clearBuildCachesDirs(ctx *buildCtx) {
	mustMkdirAll(ctx.buildRoot, 0750)
	mustMkdirAll(ctx.libdir, 0750)
	mustMkdirAll(ctx.includeDir, 0750)
	mustMkdirAll(ctx.finalRoot, 0750)
}

// =======================
// Command helpers
// =======================
func runCmd(ctx *buildCtx, dir string, name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if ctx != nil && len(ctx.env) > 0 {
		cmd.Env = ctx.env
	}
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error running command: %s %s, %v\n", name, strings.Join(args, " "), err)
		os.Exit(1)
	}
}

func outputCmd(ctx *buildCtx, dir string, name string, args ...string) []byte {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if ctx != nil && len(ctx.env) > 0 {
		cmd.Env = ctx.env
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error running command: %s %s\n%s\n%v\n",
			name, strings.Join(args, " "), string(out), err)
		os.Exit(1)
	}
	return out
}

// =======================
// FLTK source code get
// =======================
func ensureFltkSource(ctx *buildCtx) {
	stat, err := os.Stat(ctx.fltkSource)
	if errors.Is(err, fs.ErrNotExist) {
		fmt.Println("Cloning FLTK repository")
		runCmd(ctx, ctx.buildRoot, "git", "clone", "https://github.com/fltk/fltk.git")
		return
	}
	if err != nil {
		fmt.Printf("Error stating FLTK directory %s, %v\n", ctx.fltkSource, err)
		os.Exit(1)
	}
	if !stat.IsDir() {
		fmt.Printf("FLTK source path %s is not directory\n", ctx.fltkSource)
		os.Exit(1)
	}

	fmt.Println("Found existing FLTK directory")

	if ctx.goos == "windows" {
		runCmd(ctx, ctx.fltkSource, "git", "checkout", "src/Fl_win32.cxx")
	}
	runCmd(ctx, ctx.fltkSource, "git", "fetch")
}

func checkoutTargetVersion(ctx *buildCtx) {
	runCmd(ctx, ctx.fltkSource, "git", "checkout", config.FLTKPreBuildVersion)
}

func applyWindowsPatchIfNeeded(ctx *buildCtx) {
	if ctx.goos != "windows" {
		return
	}

	patchAbs, err := filepath.Abs(FLTKPatchPath)
	if err != nil {
		fmt.Printf("Error resolving patch abs path %s, %v\n", FLTKPatchPath, err)
		os.Exit(1)
	}

	if _, err := os.Stat(patchAbs); err != nil {
		fmt.Printf("Patch file not found: %s, %v\n", patchAbs, err)
		os.Exit(1)
	}

	runCmd(ctx, ctx.fltkSource, "git", "apply", patchAbs)
}

// =======================
// CMake
// =======================
func cmakeGenerator(goos string) string {
	if goos == "windows" {
		return "MinGW Makefiles"
	}
	return "Unix Makefiles"
}

func runCMakeConfigure(ctx *buildCtx) {
	args := []string{
		"-G", cmakeGenerator(ctx.goos),
		"-S", ctx.fltkSource,
		"-B", ctx.cmakeBuild,
		"-DCMAKE_BUILD_TYPE=Release",

		"-DFLTK_BUILD_TEST=OFF",
		"-DFLTK_BUILD_EXAMPLES=OFF",
		"-DFLTK_BUILD_FLUID=OFF",
		"-DFLTK_BUILD_FLTK_OPTIONS=OFF",

		// static lib
		"-DBUILD_SHARED_LIBS=OFF",
		"-DFLTK_BUILD_SHARED_LIBS=OFF",

		// bundled libs
		"-DFLTK_USE_SYSTEM_LIBJPEG=OFF",
		"-DFLTK_USE_SYSTEM_LIBPNG=OFF",
		"-DFLTK_USE_SYSTEM_ZLIB=OFF",

		// OpenGL
		"-DFLTK_USE_GL=ON",

		// Wayland
		"-DFLTK_USE_WAYLAND=OFF",

		// install to staging outputRoot
		"-DCMAKE_INSTALL_PREFIX=" + ctx.outputRoot,
		"-DCMAKE_INSTALL_INCLUDEDIR=include",
		"-DCMAKE_INSTALL_LIBDIR=" + filepath.Join("lib", ctx.goos, ctx.outArch),
	}

	if ctx.goos == "darwin" {
		args = append(args,
			"-DCMAKE_OSX_DEPLOYMENT_TARGET=12.0",
			"-DCMAKE_OSX_SYSROOT=/Library/Developer/CommandLineTools/SDKs/MacOSX.sdk",
		)
		if ctx.goarch == "amd64" {
			args = append(args, "-DCMAKE_OSX_ARCHITECTURES=x86_64")
		} else if ctx.goarch == "arm64" {
			args = append(args, "-DCMAKE_OSX_ARCHITECTURES=arm64")
		} else {
			fmt.Printf("Unsupported MacOS architecture, %s\n", ctx.goarch)
			os.Exit(1)
		}
	}

	runCmd(ctx, "", "cmake", args...)
}

func runCMakeBuild(ctx *buildCtx) {
	if ctx.goos == "windows" {
		runCmd(ctx, "", "cmake", "--build", ctx.cmakeBuild, "--verbose", "--parallel", "1")
		return
	}
	args := []string{"--build", ctx.cmakeBuild, "--parallel"}
	if ctx.goos == "openbsd" {
		args = []string{"--build", ctx.cmakeBuild}
	}
	runCmd(ctx, "", "cmake", args...)
}

func runCMakeInstall(ctx *buildCtx) {
	runCmd(ctx, "", "cmake", "--install", ctx.cmakeBuild)
}

// =======================
// post install
// =======================
func normalizeTextLF(path string) {
	b, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("Error reading %s, %v\n", path, err)
		os.Exit(1)
	}

	// CRLF -> LF
	b2 := bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n"))
	// stray CR -> LF
	b2 = bytes.ReplaceAll(b2, []byte("\r"), []byte("\n"))

	if bytes.Equal(b, b2) {
		return
	}

	if err := os.WriteFile(path, b2, 0644); err != nil {
		fmt.Printf("Error writing %s, %v\n", path, err)
		os.Exit(1)
	}
}

func moveFileCrossPlatform(src, dst string) {
	st, err := os.Stat(src)
	if err != nil {
		fmt.Printf("Error stating %s, %v\n", src, err)
		os.Exit(1)
	}
	if st.IsDir() {
		fmt.Printf("Error: %s is a directory, expected file\n", src)
		os.Exit(1)
	}

	_ = os.Remove(dst)
	mustMkdirAll(filepath.Dir(dst), 0750)

	if err := os.Rename(src, dst); err == nil {
		return
	}

	copyFile(src, dst, st.Mode().Perm())
	if err := os.Remove(src); err != nil {
		fmt.Printf("Error removing %s, %v\n", src, err)
		os.Exit(1)
	}
}

func detectFltkConfigHeader(dir string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Printf("Cannot read dir: %s, %v\n", dir, err)
		os.Exit(1)
	}

	var candidates []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		ln := strings.ToLower(name)
		if strings.Contains(ln, "config") && strings.HasSuffix(ln, ".h") {
			candidates = append(candidates, name)
		}
	}

	if len(candidates) == 0 {
		fmt.Printf("No FLTK config header found in %s\n", dir)
		os.Exit(1)
	}
	if len(candidates) > 1 {
		sort.Strings(candidates)
		fmt.Printf("Multiple possible FLTK config headers in %s: %v\n", dir, candidates)
		os.Exit(1)
	}

	return candidates[0]
}

func moveFlConfigHeader(ctx *buildCtx) {
	targetDir := filepath.Join(ctx.libdir, "FL")
	mustMkdirAll(targetDir, 0750)

	srcDir := filepath.Join(ctx.includeDir, "FL")
	cfg := detectFltkConfigHeader(srcDir)

	src := filepath.Join(srcDir, cfg)
	dst := filepath.Join(ctx.libdir, "FL", cfg)

	if _, err := os.Stat(src); err != nil {
		fmt.Printf("Error: fl_config.h not found: %s, %v\n", src, err)
		os.Exit(1)
	}

	moveFileCrossPlatform(src, dst)
	normalizeTextLF(dst)

	if _, err := os.Stat(dst); err != nil {
		fmt.Printf("Error: fl_config.h not created: %s, %v\n", dst, err)
		os.Exit(1)
	}
}

func ensureFlags(s string, flags ...string) string {
	s = " " + strings.TrimSpace(s) + " "

	for _, f := range flags {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		pat := " " + f + " "
		if !strings.Contains(s, pat) {
			s += f + " "
		}
	}
	return strings.TrimSpace(s)
}

func dedupBySpace(s string) string {
	fields := strings.Fields(s)
	seen := make(map[string]bool, len(fields))
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		if !seen[f] {
			seen[f] = true
			out = append(out, f)
		}
	}
	return strings.Join(out, " ")
}

// =======================
// CGO
// =======================
func generateCgo(ctx *buildCtx) {
	mustMkdirAll(CGOOutDir, 0750)

	// darwin universal name set
	suffix := ctx.goarch
	if ctx.goos == "darwin" && ctx.outArch == "universal" {
		suffix = "universal"
	}
	name := fmt.Sprintf("cgo_%s_%s.go", ctx.goos, suffix)
	outPath := filepath.Join(CGOOutDir, name)

	f, err := os.Create(outPath)
	if err != nil {
		fmt.Printf("Error creating %s, %v\n", outPath, err)
		os.Exit(1)
	}
	defer f.Close()

	// build tag
	isDarwinUniversal := (ctx.goos == "darwin" && ctx.outArch == "universal")
	if isDarwinUniversal {
		fmt.Fprintln(f, "//go:build darwin && (amd64 || arm64)\n")
	} else {
		fmt.Fprintf(f, "//go:build %s && %s\n\n", ctx.goos, ctx.goarch)
	}
	fmt.Fprintf(f, "package %s\n\n", CGOPackage)

	// universal #cgo handle
	cgoCondOSArch := fmt.Sprintf("%s,%s", ctx.goos, ctx.goarch)
	if isDarwinUniversal {
		cgoCondOSArch = "darwin"
	}

	// 1) get parms from fltk-config
	cxx := strings.TrimSpace(runFltkConfig(ctx, "--use-gl", "--use-images", "--use-forms", "--cxxflags"))
	ld := strings.TrimSpace(runFltkConfig(ctx, "--use-gl", "--use-images", "--use-forms", "--ldstaticflags"))

	// 2) rewrite final dir（libs/fltk）
	cxx = strings.TrimSpace(rewritePathsForCgo(ctx, cxx))
	ld = strings.TrimSpace(rewritePathsForCgo(ctx, ld))

	// 3) macOS：weak_framework replace -framework
	if ctx.goos == "darwin" {
		ld = strings.ReplaceAll(ld, "-weak_framework UniformTypeIdentifiers", "-framework UniformTypeIdentifiers")
	}

	// 4) largefile
	if ctx.goos == "windows" || ctx.goos == "linux" {
		if !strings.Contains(cxx, "-D_LARGEFILE_SOURCE") {
			cxx += " -D_LARGEFILE_SOURCE"
		}
		if !strings.Contains(cxx, "-D_LARGEFILE64_SOURCE") {
			cxx += " -D_LARGEFILE64_SOURCE"
		}
		if !strings.Contains(cxx, "-D_FILE_OFFSET_BITS=64") {
			cxx += " -D_FILE_OFFSET_BITS=64"
		}
	}

	// 5)  fl_config.h
	finalRoot := "${SRCDIR}/../" + filepath.ToSlash(ctx.finalRoot)
	incArch := fmt.Sprintf("-I%s/%s/%s", finalRoot, ctx.goos, ctx.outArch) // 最前（命中 FL/fl_config.h）
	incBase := fmt.Sprintf("-I%s/include", finalRoot)
	incImg := fmt.Sprintf("-I%s/include/FL/images", finalRoot)

	rmSet := map[string]struct{}{
		incArch: {},
		incBase: {},
		incImg:  {},
	}

	toks := strings.Fields(cxx)
	kept := make([]string, 0, len(toks))
	seen := make(map[string]struct{}, len(toks))

	for _, t := range toks {
		if _, ok := rmSet[t]; ok {
			continue
		}

		if strings.HasPrefix(t, "/") && strings.Contains(t, "FL/images") && !strings.HasPrefix(t, "-I") {
			continue
		}

		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		kept = append(kept, t)
	}
	cxx = strings.Join(kept, " ")

	// fl_config.h -> incArch -> incBase -> incImg -> cxx(other flags)
	cppflags := strings.TrimSpace(strings.Join([]string{incArch, incBase, incImg, cxx}, " "))

	if ctx.goos == "windows" {
		ld = removeTokens(ld, "-lfontconfig")

		if !strings.Contains(ld, "-mwindows") {
			ld = "-mwindows " + ld
		}
	}

	// output #cgo
	fmt.Fprintf(f, "// #cgo %s CPPFLAGS: %s\n", cgoCondOSArch, cppflags)
	fmt.Fprintf(f, "// #cgo %s CXXFLAGS: -std=%s\n", cgoCondOSArch, config.FLTKCppStandard)
	fmt.Fprintf(f, "// #cgo %s LDFLAGS: %s\n", cgoCondOSArch, strings.TrimSpace(ld))
	fmt.Fprintln(f, `import "C"`)

	fmt.Printf("Generated cgo file: %s\n", outPath)
}

func removeTokens(s string, bad ...string) string {
	rm := make(map[string]struct{}, len(bad))
	for _, b := range bad {
		rm[strings.TrimSpace(b)] = struct{}{}
	}
	toks := strings.Fields(s)
	out := make([]string, 0, len(toks))
	for _, t := range toks {
		if _, ok := rm[t]; ok {
			continue
		}
		out = append(out, t)
	}
	return strings.Join(out, " ")
}

// =======================
// FS helpers
// =======================
func mustRemoveAll(path string) {
	_ = os.RemoveAll(path)
}

func copyFile(src, dst string, perm fs.FileMode) {
	in, err := os.Open(src)
	if err != nil {
		fmt.Printf("Error opening %s, %v\n", src, err)
		os.Exit(1)
	}
	defer in.Close()

	mustMkdirAll(filepath.Dir(dst), 0750)

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, perm)
	if err != nil {
		fmt.Printf("Error creating %s, %v\n", dst, err)
		os.Exit(1)
	}
	defer out.Close()

	if _, err := out.ReadFrom(in); err != nil {
		fmt.Printf("Error copying %s -> %s, %v\n", src, dst, err)
		os.Exit(1)
	}
}

func copyDir(srcDir, dstDir string) {
	err := filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(srcDir, path)
		dst := filepath.Join(dstDir, rel)

		if d.IsDir() {
			return os.MkdirAll(dst, 0750)
		}

		info, err := d.Info()
		if err != nil {
			return err
		}
		copyFile(path, dst, info.Mode().Perm())
		return nil
	})
	if err != nil {
		fmt.Printf("Error copying dir %s -> %s, %v\n", srcDir, dstDir, err)
		os.Exit(1)
	}
}

func syncArtifactsToFinalRoot(ctx *buildCtx) {
	// 1) include -> libs/fltk/include
	srcInclude := filepath.Join(ctx.outputRoot, "include")
	dstInclude := filepath.Join(ctx.finalRoot, "include")
	mustRemoveAll(dstInclude)
	copyDir(srcInclude, dstInclude)

	// 2) lib/<os>/<arch> -> libs/fltk/<os>/<arch>
	srcLib := filepath.Join(ctx.outputRoot, "lib", ctx.goos, ctx.outArch)
	dstLib := filepath.Join(ctx.finalRoot, ctx.goos, ctx.outArch)
	mustRemoveAll(dstLib)
	copyDir(srcLib, dstLib)
}

func cleanStagingInstall(ctx *buildCtx) {
	_ = os.RemoveAll(ctx.outputRoot)
}

func mustFileExist(path string) {
	if _, err := os.Stat(path); err != nil {
		fmt.Printf("Missing required file: %s, %v\n", path, err)
		os.Exit(1)
	}
}

func mustChmodX(path string) {
	st, err := os.Stat(path)
	if err != nil {
		fmt.Printf("Missing required file: %s, %v\n", path, err)
		os.Exit(1)
	}
	if runtime.GOOS != "windows" {
		_ = os.Chmod(path, st.Mode().Perm()|0111)
	}
}

func fltkConfigPath(ctx *buildCtx) string {
	p := filepath.Join(ctx.cmakeBuild, "bin", "fltk-config")

	if ctx.goos == "darwin" && ctx.outArch == "universal" {
		candidates := []string{
			filepath.Join(ctx.buildRoot, "fltk-cmake-darwin-arm64", "bin", "fltk-config"),
			filepath.Join(ctx.buildRoot, "fltk-cmake-darwin-amd64", "bin", "fltk-config"),
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				return c
			}
		}
		return p
	}

	return p
}

func findGitShPath() (string, error) {
	possiblePaths := []string{
		filepath.Join(os.Getenv("ProgramFiles"), "Git", "bin", "sh.exe"),
		filepath.Join(os.Getenv("ProgramFiles(x86)"), "Git", "bin", "sh.exe"),
		filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local", "GitForWindows", "bin", "sh.exe"),
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	if gitHome := os.Getenv("GIT_HOME"); gitHome != "" {
		path := filepath.Join(gitHome, "bin", "sh.exe")
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", errors.New("sh.exe not found in common Git paths")
}

func runFltkConfig(ctx *buildCtx, args ...string) string {
	cfg := fltkConfigPath(ctx)
	mustChmodX(cfg)

	if ctx.goos == "windows" {
		shPath, err := findGitShPath()
		if err != nil {
			fmt.Printf("Failed to find sh.exe: %v, fallback to 'sh'\n", err)
			shPath = "sh"
		}
		cmdArgs := append([]string{"-c", fmt.Sprintf("\"%s\" %s", cfg, strings.Join(args, " "))})
		out := outputCmd(ctx, "", shPath, cmdArgs...)
		return string(out)
	}
	out := outputCmd(ctx, "", cfg, args...)
	return string(out)
}

func rewritePathsForCgo(ctx *buildCtx, s string) string {
	root := "${SRCDIR}/../" + filepath.ToSlash(ctx.finalRoot)

	// fix windows path_url
	ss := strings.ReplaceAll(s, "\\", "/")

	outRel := filepath.ToSlash(filepath.Clean(ctx.outputRoot))

	outAbs, _ := filepath.Abs(ctx.outputRoot)
	outAbs = filepath.ToSlash(filepath.Clean(outAbs))

	incAbs, _ := filepath.Abs(filepath.Join(ctx.outputRoot, "include"))
	incAbs = filepath.ToSlash(filepath.Clean(incAbs))

	libAbs, _ := filepath.Abs(filepath.Join(ctx.outputRoot, "lib", ctx.goos, ctx.outArch))
	libAbs = filepath.ToSlash(filepath.Clean(libAbs))

	wdAbs := filepath.ToSlash(filepath.Clean(ctx.currentDir))

	repls := []string{
		libAbs, root + "/" + filepath.ToSlash(filepath.Join(ctx.goos, ctx.outArch)),
		incAbs, root + "/include",

		outAbs, root,
		outRel, root,

		wdAbs, root,
	}

	for i := 0; i+1 < len(repls); i += 2 {
		ss = strings.ReplaceAll(ss, repls[i], repls[i+1])
	}

	ss = strings.TrimSpace(ss)

	if ctx.goos == "darwin" && ctx.outArch == "universal" {
		ss = strings.ReplaceAll(ss, "/libs/fltk/darwin/arm64/", "/libs/fltk/darwin/universal/")
		ss = strings.ReplaceAll(ss, "/libs/fltk/darwin/amd64/", "/libs/fltk/darwin/universal/")
		ss = strings.ReplaceAll(ss, "/libs/fltk/lib/darwin/arm64/", "/libs/fltk/darwin/universal/")
		ss = strings.ReplaceAll(ss, "/libs/fltk/lib/darwin/amd64/", "/libs/fltk/darwin/universal/")
	}

	return ss

}

func cleanCMakeBuildDir(ctx *buildCtx) {
	_ = os.RemoveAll(ctx.cmakeBuild)
}

// =======================
// macOS universal merge
// =======================
func mergeDarwinUniversal(ctx *buildCtx) {
	if ctx.goos != "darwin" || ctx.outArch != "universal" {
		return
	}

	fmt.Println("Merging macOS universal binaries (lipo)")

	base := filepath.Join(ctx.finalRoot, "darwin")
	arm64Dir := filepath.Join(base, "arm64")
	amd64Dir := filepath.Join(base, "amd64")
	universalDir := filepath.Join(base, "universal")

	if _, err := os.Stat(arm64Dir); err != nil {
		fmt.Printf("arm64 libs not found: %s\n", arm64Dir)
		os.Exit(1)
	}
	if _, err := os.Stat(amd64Dir); err != nil {
		fmt.Printf("amd64 libs not found: %s\n", amd64Dir)
		os.Exit(1)
	}

	mustMkdirAll(universalDir, 0750)
	mustMkdirAll(filepath.Join(universalDir, "FL"), 0750)

	entries, err := os.ReadDir(arm64Dir)
	if err != nil {
		fmt.Printf("Error reading %s, %v\n", arm64Dir, err)
		os.Exit(1)
	}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".a") {
			continue
		}
		armLib := filepath.Join(arm64Dir, e.Name())
		amdLib := filepath.Join(amd64Dir, e.Name())
		outLib := filepath.Join(universalDir, e.Name())

		if _, err := os.Stat(amdLib); err != nil {
			fmt.Printf("Missing amd64 lib for %s\n", e.Name())
			os.Exit(1)
		}

		runCmd(ctx, "", "lipo", "-create", armLib, amdLib, "-output", outLib)
	}

	srcCfg := filepath.Join(arm64Dir, "FL", "fl_config.h")
	dstCfg := filepath.Join(universalDir, "FL", "fl_config.h")
	copyFile(srcCfg, dstCfg, 0644)
	normalizeTextLF(dstCfg)

	mustFileExist(filepath.Join(universalDir, "libfltk.a"))

	_ = os.RemoveAll(arm64Dir)
	_ = os.RemoveAll(amd64Dir)

	fmt.Println("Cleaned darwin thin libs: arm64/amd64 removed, kept universal only")
	fmt.Println("macOS universal merge completed")
}

// =======================
// main process
// =======================
func buildStaticLibs(ctx *buildCtx) {
	clearBuildCachesDirs(ctx)

	ensureFltkSource(ctx)
	checkoutTargetVersion(ctx)
	applyWindowsPatchIfNeeded(ctx)

	cleanCMakeBuildDir(ctx)
	runCMakeConfigure(ctx)
	runCMakeBuild(ctx)
	cleanStagingInstall(ctx)
	runCMakeInstall(ctx)

	// fl_config.h from include/FL -to> staging <os>/<arch>/FL
	moveFlConfigHeader(ctx)

	// ✅ staging -> final: libs/fltk/include + libs/fltk/lib/<os>/<arch>
	syncArtifactsToFinalRoot(ctx)

	// ✅ check final product
	mustFileExist(filepath.Join(ctx.finalRoot, ctx.goos, ctx.outArch, "FL", "fl_config.h"))
	mustFileExist(filepath.Join(ctx.finalRoot, ctx.goos, ctx.outArch, "libfltk.a"))

	writeManifestForTarget(ctx, ctx.goos, ctx.outArch)
}

func main() {
	mustCheckEnv()
	mustCheckTool("git")
	mustCheckTool("cmake")

	if runtime.GOOS == "darwin" {
		buildStaticLibs(newBuildCtx("arm64", "arm64"))
		buildStaticLibs(newBuildCtx("amd64", "amd64"))

		ctxUni := newBuildCtx("arm64", "universal")
		mergeDarwinUniversal(ctxUni)

		// manifest + cgo
		writeManifestForTarget(ctxUni, "darwin", "universal")
		generateCgo(ctxUni)

		fmt.Println("Successfully built darwin arm64 + amd64 + universal")
		return
	}

	// other arch
	ctx := newBuildCtx(runtime.GOARCH, runtime.GOARCH)
	buildStaticLibs(ctx)
	generateCgo(ctx)

	fmt.Printf("Successfully generated libraries for OS: %s, architecture: %s\n", ctx.goos, ctx.goarch)
}
