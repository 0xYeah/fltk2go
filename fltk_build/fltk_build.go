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

// =======================
// 全局可配置变量
// =======================
var (
	// 构建/缓存根目录（强烈建议统一放 build/，不提交）
	// 默认: build
	FLTKBuildRoot = getEnvOrDefault("FLTK_BUILD_ROOT", "build")

	// 成品输出根目录（include / lib / cgo_* 文件）
	// 默认: 当前项目根目录
	OutputRoot = getEnvOrDefault("FLTK_OUTPUT_ROOT", filepath.Join("build", "_install"))

	// 仅用于输出目录命名（lib/<os>/<arch>）
	// 例如 darwin 想输出到 lib/darwin/universal：设置 FLTK_OUTPUT_ARCH=universal
	OutputArchOverride = os.Getenv("FLTK_OUTPUT_ARCH")

	// cgo 文件输出目录（新的绑定层包目录）
	// 默认：fltk_bridge
	CGOOutDir = getEnvOrDefault("FLTK_CGO_OUT_DIR", "fltk_bridge")

	// cgo 文件 package 名（默认为 fltk_bridge）
	CGOPackage = getEnvOrDefault("FLTK_CGO_PACKAGE", "fltk_bridge")

	// Windows patch 路径（脚本资产，应该提交；不要放进 build 缓存）
	// 默认: fltk_build/fltk-1.4.patch
	FLTKPatchPath = getEnvOrDefault("FLTK_PATCH_PATH", filepath.Join("fltk_build", "fltk-1.4.patch"))
)

func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// =======================
// 构建上下文
// =======================
type buildCtx struct {
	goos    string
	goarch  string
	outArch string

	// 输出相关
	outputRoot string
	libdir     string
	includeDir string

	// build 缓存相关
	buildRoot  string
	fltkSource string
	cmakeBuild string

	currentDir string
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

func newBuildCtx() *buildCtx {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	outputRoot := filepath.Clean(OutputRoot)
	buildRoot := filepath.Clean(FLTKBuildRoot)

	outArch := goarch
	if OutputArchOverride != "" {
		outArch = OutputArchOverride
	}

	libdir := filepath.Join(outputRoot, "lib", goos, outArch)
	includeDir := filepath.Join(outputRoot, "include")

	aCtx := &buildCtx{
		goos:       goos,
		goarch:     goarch,
		outArch:    outArch,
		outputRoot: outputRoot,
		libdir:     libdir,
		includeDir: includeDir,

		buildRoot:  buildRoot,
		fltkSource: filepath.Join(buildRoot, "fltk"),
		cmakeBuild: filepath.Join(buildRoot, "fltk-cmake"),

		currentDir: mustGetwd(),
	}
	return aCtx
}

func prepareDirs(ctx *buildCtx) {
	mustMkdirAll(ctx.buildRoot, 0750)
	mustMkdirAll(ctx.libdir, 0750)
	mustMkdirAll(ctx.includeDir, 0750)
}

// =======================
// Command helpers
// =======================
func runCmd(dir string, name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error running command: %s %s, %v\n", name, strings.Join(args, " "), err)
		os.Exit(1)
	}
}

func outputCmd(dir string, name string, args ...string) []byte {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error running command: %s %s, %v\n", name, strings.Join(args, " "), err)
		os.Exit(1)
	}
	return out
}

// =======================
// FLTK 源码准备
// =======================
func ensureFltkSource(ctx *buildCtx) {
	stat, err := os.Stat(ctx.fltkSource)
	if errors.Is(err, fs.ErrNotExist) {
		fmt.Println("Cloning FLTK repository")
		runCmd(ctx.buildRoot, "git", "clone", "https://github.com/fltk/fltk.git")
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
		runCmd(ctx.fltkSource, "git", "checkout", "src/Fl_win32.cxx")
	}
	runCmd(ctx.fltkSource, "git", "fetch")
}

func checkoutTargetVersion(ctx *buildCtx) {
	runCmd(ctx.fltkSource, "git", "checkout", config.FLTKPreBuildVersion)
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

	runCmd(ctx.fltkSource, "git", "apply", patchAbs)
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

		// 静态库
		"-DBUILD_SHARED_LIBS=OFF",

		// install 到输出根目录；libdir 仍保持 lib/<os>/<arch> 结构
		"-DCMAKE_INSTALL_PREFIX=" + ctx.outputRoot,
		"-DCMAKE_INSTALL_INCLUDEDIR=include",
		"-DCMAKE_INSTALL_LIBDIR=" + filepath.Join("lib", ctx.goos, ctx.outArch),
	}

	// 这些选项在 Windows 上通常不会被 FLTK 使用（会触发 unused warning），所以仅在非 Windows 传
	if ctx.goos != "windows" {
		args = append(args,
			"-DFLTK_USE_SYSTEM_JPEG=OFF",
			"-DFLTK_USE_SYSTEM_PNG=OFF",
			"-DFLTK_USE_SYSTEM_ZLIB=OFF",
		)
	}

	// Wayland 仅 Linux 有意义
	if ctx.goos == "linux" {
		args = append(args, "-DFLTK_USE_WAYLAND=OFF")
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

	runCmd("", "cmake", args...)
}

func runCMakeBuild(ctx *buildCtx) {
	args := []string{"--build", ctx.cmakeBuild, "--parallel"}
	if ctx.goos == "openbsd" {
		args = []string{"--build", ctx.cmakeBuild}
	}
	runCmd("", "cmake", args...)
}

func runCMakeInstall(ctx *buildCtx) {
	runCmd("", "cmake", "--install", ctx.cmakeBuild)
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

	// 内容没变就不写回
	if bytes.Equal(b, b2) {
		return
	}

	if err := os.WriteFile(path, b2, 0644); err != nil {
		fmt.Printf("Error writing %s, %v\n", path, err)
		os.Exit(1)
	}
}

func moveFileCrossPlatform(src, dst string) {
	// 源必须存在
	st, err := os.Stat(src)
	if err != nil {
		fmt.Printf("Error stating %s, %v\n", src, err)
		os.Exit(1)
	}
	if st.IsDir() {
		fmt.Printf("Error: %s is a directory, expected file\n", src)
		os.Exit(1)
	}

	// 目标若存在，明确覆盖：先删
	_ = os.Remove(dst)

	// 确保目标目录存在
	mustMkdirAll(filepath.Dir(dst), 0750)

	// 先尝试 Rename（同盘最快）
	if err := os.Rename(src, dst); err == nil {
		return
	}

	// Rename 失败则 copy + remove
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

	// 源文件必须存在（否则 install 没产出，直接 fail-fast）
	if _, err := os.Stat(src); err != nil {
		fmt.Printf("Error: fl_config.h not found: %s, %v\n", src, err)
		os.Exit(1)
	}

	moveFileCrossPlatform(src, dst)

	// 核心需求：统一 LF
	normalizeTextLF(dst)

	// 可选：再确认 dst 存在
	if _, err := os.Stat(dst); err != nil {
		fmt.Printf("Error: fl_config.h not created: %s, %v\n", dst, err)
		os.Exit(1)
	}
}

// =======================
// CGO
// =======================
func listStaticLibs(relLibArch string) []string {
	dir := filepath.FromSlash(relLibArch)
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Printf("Error reading %s, %v\n", dir, err)
		os.Exit(1)
	}

	var libs []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasPrefix(name, "libfltk") && strings.HasSuffix(name, ".a") {
			libs = append(libs, "${SRCDIR}/"+filepath.ToSlash(filepath.Join(relLibArch, name)))
		}
	}

	if len(libs) == 0 {
		fmt.Printf("No libfltk*.a found in %s\n", dir)
		os.Exit(1)
	}

	return libs
}

func generateCgo(ctx *buildCtx) {
	// 确保目录存在
	mustMkdirAll(CGOOutDir, 0750)

	// 文件名：写到 fltk_bridge/ 下
	name := fmt.Sprintf("cgo_%s_%s.go", ctx.goos, ctx.goarch)
	outPath := filepath.Join(CGOOutDir, name)

	f, err := os.Create(outPath)
	if err != nil {
		fmt.Printf("Error creating %s, %v\n", outPath, err)
		os.Exit(1)
	}
	defer f.Close()

	fmt.Fprintf(f, "//go:build %s && %s\n\n", ctx.goos, ctx.goarch)
	fmt.Fprintf(f, "package %s\n\n", CGOPackage)

	// 注意：SRCDIR 现在是 fltk_bridge 目录，所以要 ../lib/...
	relLibArch := filepath.ToSlash(filepath.Join("..", "lib", ctx.goos, ctx.outArch))
	relLibInclude := filepath.ToSlash(filepath.Join("..", "lib", "include"))

	largeFileDefines := ""
	if ctx.goos == "windows" || ctx.goos == "linux" {
		largeFileDefines = " -D_LARGEFILE_SOURCE -D_LARGEFILE64_SOURCE -D_FILE_OFFSET_BITS=64"
	}

	// CPPFLAGS：顺序不变（先平台 arch，命中 fl_config.h）
	fmt.Fprintf(
		f,
		"// #cgo %s,%s CPPFLAGS: -I${SRCDIR}/%s -I${SRCDIR}/%s -I${SRCDIR}/%s/FL/images%s\n",
		ctx.goos, ctx.goarch,
		relLibArch,
		relLibInclude,
		relLibInclude,
		largeFileDefines,
	)

	// CXXFLAGS：保持旧版
	fmt.Fprintf(f, "// #cgo %s,%s CXXFLAGS: -std=c++11\n", ctx.goos, ctx.goarch)

	// 静态库：你原来那套固定顺序（这里用新的 relLibArch）
	libs := orderedFltkStaticLibs(filepath.ToSlash(filepath.Join("lib", ctx.goos, ctx.outArch)))
	// ↑ orderedFltkStaticLibs 期望的是相对“项目根”的 lib/<os>/<arch>，
	//   它内部用 os.ReadDir(baseDir) 读取实际目录，所以传 "lib/..." 不要带 ".."。
	libFlags := strings.Join(libs, " ")

	switch ctx.goos {
	case "windows":
		fmt.Fprintf(
			f,
			"// #cgo %s,%s LDFLAGS: -mwindows %s -lglu32 -lopengl32 -lgdiplus -lole32 -luuid -lcomctl32 -lws2_32 -lwinspool\n",
			ctx.goos, ctx.goarch, libFlags,
		)

	case "darwin":
		fmt.Fprintf(
			f,
			"// #cgo %s,%s LDFLAGS: %s -framework Cocoa -framework Carbon -framework ApplicationServices -framework CoreGraphics -framework CoreText -framework QuartzCore -framework IOKit -framework OpenGL\n",
			ctx.goos, ctx.goarch, libFlags,
		)

	case "linux":
		fmt.Fprintf(
			f,
			"// #cgo %s,%s LDFLAGS: %s -lX11 -lXext -lXinerama -lXcursor -lXrender -lXfixes -lXft -lfontconfig -lfreetype -lGL -lGLU -ldl -lpthread -lm\n",
			ctx.goos, ctx.goarch, libFlags,
		)

	default:
		fmt.Fprintf(f, "// #cgo %s,%s LDFLAGS: %s\n", ctx.goos, ctx.goarch, libFlags)
	}

	fmt.Fprintln(f, `import "C"`)

	fmt.Printf("Generated cgo file: %s\n", outPath)
}

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

func syncArtifactsToProjectRoot(ctx *buildCtx) {
	// 目标：项目根（固定）
	finalRoot := projectRoot()

	// 1) include
	srcInclude := filepath.Join(ctx.outputRoot, "include")
	dstInclude := filepath.Join(finalRoot, "lib", "include")
	mustRemoveAll(dstInclude)
	copyDir(srcInclude, dstInclude)

	// 2) lib/<os>/<arch>
	srcLib := filepath.Join(ctx.outputRoot, "lib", ctx.goos, ctx.outArch)
	dstLib := filepath.Join(finalRoot, "lib", ctx.goos, ctx.outArch)
	mustRemoveAll(dstLib)
	copyDir(srcLib, dstLib)
}

func mergeDarwinUniversal(ctx *buildCtx) {
	if ctx.goos != "darwin" || ctx.outArch != "universal" {
		return
	}

	fmt.Println("Merging macOS universal binaries (lipo)")

	base := filepath.Join("lib", "darwin")
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

		runCmd("", "lipo",
			"-create",
			armLib,
			amdLib,
			"-output",
			outLib,
		)
	}

	// fl_config.h：任选一个（arm64 / amd64 内容几乎一致）
	srcCfg := filepath.Join(arm64Dir, "FL", "fl_config.h")
	dstCfg := filepath.Join(universalDir, "FL", "fl_config.h")
	copyFile(srcCfg, dstCfg, 0644)
	normalizeTextLF(dstCfg)
	fmt.Println("macOS universal merge completed")
}

func orderedFltkStaticLibs(relLibArch string) []string {
	// 旧版固定顺序（你原先手写的那套）
	ordered := []string{
		"libfltk_images.a",
		"libfltk_jpeg.a",
		"libfltk_png.a",
		"libfltk_z.a",
		"libfltk_gl.a",
		"libfltk_forms.a",
		"libfltk.a",
	}

	baseDir := filepath.FromSlash(relLibArch)

	// 先收集目录里所有 libfltk*.a（用于兜底追加）
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		fmt.Printf("Error reading %s, %v\n", baseDir, err)
		os.Exit(1)
	}
	allSet := map[string]bool{}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		n := e.Name()
		if strings.HasPrefix(n, "libfltk") && strings.HasSuffix(n, ".a") {
			allSet[n] = true
		}
	}

	// 按固定顺序取；缺任何一个直接报错（避免静默生成坏 cgo）
	var libs []string
	for _, name := range ordered {
		if !allSet[name] {
			fmt.Printf("Missing required static lib: %s (dir: %s)\n", name, baseDir)
			os.Exit(1)
		}
		libs = append(libs, "${SRCDIR}/"+filepath.ToSlash(filepath.Join(relLibArch, name)))
		delete(allSet, name)
	}

	// 兜底：如果未来 FLTK 新增 libfltk_xxx.a，追加到末尾（稳定且不破坏旧顺序）
	if len(allSet) > 0 {
		var extra []string
		for name := range allSet {
			extra = append(extra, name)
		}
		sort.Strings(extra)
		for _, name := range extra {
			libs = append(libs, "${SRCDIR}/"+filepath.ToSlash(filepath.Join(relLibArch, name)))
		}
	}

	return libs
}

func cleanStagingInstall(ctx *buildCtx) {
	// 只清理本次目标会写入的路径，避免误删其他平台的产物
	_ = os.RemoveAll(filepath.Join(ctx.outputRoot, "include"))
	_ = os.RemoveAll(filepath.Join(ctx.outputRoot, "lib", ctx.goos, ctx.outArch))
}

func mustFileExist(path string) {
	if _, err := os.Stat(path); err != nil {
		fmt.Printf("Missing required file: %s, %v\n", path, err)
		os.Exit(1)
	}
}

func projectRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println("cannot get working directory:", err)
		os.Exit(1)
	}

	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			if _, err := os.Stat(filepath.Join(wd, "fltk_bridge")); err == nil {
				return wd
			}
		}

		parent := filepath.Dir(wd)
		if parent == wd {
			fmt.Println("cannot locate fltk2go project root")
			os.Exit(1)
		}
		wd = parent
	}
}

// =======================
// main
// =======================
func main() {
	mustCheckEnv()
	mustCheckTool("git")
	mustCheckTool("cmake")

	root := projectRoot()
	_ = os.Chdir(root) // ⭐ 核心修复点：统一工作目录

	ctx := newBuildCtx()

	fmt.Printf("Build root (cache): %s\n", ctx.buildRoot)
	fmt.Printf("Output root       : %s\n", ctx.outputRoot)
	fmt.Printf("Lib output dir    : %s\n", ctx.libdir)

	prepareDirs(ctx)

	ensureFltkSource(ctx)
	checkoutTargetVersion(ctx)
	applyWindowsPatchIfNeeded(ctx)

	runCMakeConfigure(ctx)
	runCMakeBuild(ctx)
	// 在 install 之前先清 staging，确保 cmake --install 输出没有旧文件残留
	cleanStagingInstall(ctx)
	runCMakeInstall(ctx)

	// staging 内处理 fl_config.h
	moveFlConfigHeader(ctx)

	// 同步 arm64 / amd64 / staging 产物到项目根 lib/
	syncArtifactsToProjectRoot(ctx)

	// 写当前 target 的 manifest（写到项目根的 lib/<os>/<arch>）
	writeManifestForTarget(ctx, ctx.goos, ctx.outArch)

	// ⭐️ 只有 darwin + universal 时，这一步才真正执行
	mergeDarwinUniversal(ctx)
	if ctx.goos == "darwin" && ctx.outArch == "universal" {
		writeManifestForTarget(ctx, "darwin", "universal")
	}

	// 同步后确认 fl_config.h 确实在最终位置
	mustFileExist(filepath.Join("lib", ctx.goos, ctx.outArch, "FL", "fl_config.h"))

	// cgo 永远基于最终 lib/<os>/<arch> 或 lib/<os>/universal
	generateCgo(ctx)

	fmt.Printf("Successfully generated libraries for OS: %s, architecture: %s\n",
		ctx.goos, ctx.goarch)
}
