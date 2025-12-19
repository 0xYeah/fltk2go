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
func newBuildCtx(goarch, outArch string) *buildCtx {
	goos := runtime.GOOS

	outputRoot := filepath.Clean(OutputRoot)
	buildRoot := filepath.Clean(FLTKBuildRoot)

	libdir := filepath.Join(outputRoot, "lib", goos, outArch)
	includeDir := filepath.Join(outputRoot, "include")

	return &buildCtx{
		goos:       goos,
		goarch:     goarch,
		outArch:    outArch,
		outputRoot: outputRoot,
		libdir:     libdir,
		includeDir: includeDir,

		buildRoot:  buildRoot,
		fltkSource: filepath.Join(buildRoot, "fltk"),
		// 必须按 arch 隔离，否则一定踩坑
		cmakeBuild: filepath.Join(buildRoot, "fltk-cmake-"+outArch),

		currentDir: mustGetwd(),
	}
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

		"-DOPTION_USE_WAYLAND=OFF",
		"-DOPTION_USE_SYSTEM_LIBJPEG=OFF",
		"-DOPTION_USE_SYSTEM_LIBPNG=OFF",
		"-DOPTION_USE_SYSTEM_ZLIB=OFF",

		// install 到输出根目录；libdir 仍保持 lib/<os>/<arch> 结构
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
	mustMkdirAll(CGOOutDir, 0750)

	// 文件名
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

	// SRCDIR 在 fltk_bridge 目录，所以 lib/include 都要 ../
	relLibArchForCgo := filepath.ToSlash(filepath.Join("..", "lib", ctx.goos, ctx.outArch))
	relIncludeForCgo := filepath.ToSlash(filepath.Join("..", "lib", "include"))

	largeFileDefines := ""
	if ctx.goos == "windows" || ctx.goos == "linux" {
		largeFileDefines = " -D_LARGEFILE_SOURCE -D_LARGEFILE64_SOURCE -D_FILE_OFFSET_BITS=64"
	}

	// 关键：universal 的 #cgo 条件必须用 "darwin"（不带 arch），否则 amd64 不生效
	cgoCondOSArch := fmt.Sprintf("%s,%s", ctx.goos, ctx.goarch)
	if isDarwinUniversal {
		cgoCondOSArch = "darwin"
	}

	// CPPFLAGS：先 arch，再 include，再 images
	fmt.Fprintf(
		f,
		"// #cgo %s CPPFLAGS: -I${SRCDIR}/%s -I${SRCDIR}/%s -I${SRCDIR}/%s/FL/images%s\n",
		cgoCondOSArch,
		relLibArchForCgo,
		relIncludeForCgo,
		relIncludeForCgo,
		largeFileDefines,
	)

	fmt.Fprintf(f, "// #cgo %s CXXFLAGS: -std=%s\n", cgoCondOSArch, config.FLTKCppStandard)

	// 注意：检查目录用“项目根下的真实路径”，输出用“相对 SRCDIR 的路径”
	checkDir := filepath.ToSlash(filepath.Join("lib", ctx.goos, ctx.outArch))      // 不带 ..
	emitRel := filepath.ToSlash(filepath.Join("..", "lib", ctx.goos, ctx.outArch)) // 只允许相对片段

	libs := orderedFltkStaticLibs(checkDir, emitRel)

	// 保险：强制禁止绝对路径进入 LDFLAGS（避免 fltk-config 输出混进来）
	for _, x := range libs {
		if strings.Contains(x, "/Users/") || strings.Contains(x, ":\\") || strings.HasPrefix(x, "/") {
			fmt.Printf("Error: absolute path detected in static libs: %s\n", x)
			os.Exit(1)
		}
	}
	libFlags := strings.Join(libs, " ")

	switch ctx.goos {
	case "windows":
		fmt.Fprintf(
			f,
			"// #cgo %s LDFLAGS: -mwindows %s -lglu32 -lopengl32 -lgdiplus -lole32 -luuid -lcomctl32 -lws2_32 -lwinspool\n",
			cgoCondOSArch, libFlags,
		)

	case "darwin":
		fmt.Fprintf(
			f,
			"// #cgo %s LDFLAGS: %s -lm -lpthread -framework Cocoa -framework OpenGL -framework UniformTypeIdentifiers\n",
			cgoCondOSArch, libFlags,
		)

	case "linux":
		fmt.Fprintf(
			f,
			"// #cgo %s LDFLAGS: %s -lX11 -lXext -lXinerama -lXcursor -lXrender -lXfixes -lXft -lfontconfig -lfreetype -lGL -lGLU -ldl -lpthread -lm\n",
			cgoCondOSArch, libFlags,
		)

	default:
		fmt.Fprintf(f, "// #cgo %s LDFLAGS: %s\n", cgoCondOSArch, libFlags)
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
	finalRoot := "."

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

	mustFileExist(filepath.Join(universalDir, "libfltk.a"))
	_ = os.RemoveAll(arm64Dir)
	_ = os.RemoveAll(amd64Dir)
	fmt.Println("Cleaned darwin thin libs: arm64/amd64 removed, kept universal only")

	fmt.Println("macOS universal merge completed")
}

func orderedFltkStaticLibs(checkDirFromProjectRoot, emitRelFromSRCDIR string) []string {
	ordered := []string{
		"libfltk_images.a",
		"libfltk_jpeg.a",
		"libfltk_png.a",
		"libfltk_z.a",
		"libfltk_gl.a",
		"libfltk_forms.a",
		"libfltk.a",
	}

	// 1) 检查用：项目根下的真实目录（不带 ..）
	baseDir := filepath.FromSlash(checkDirFromProjectRoot)

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

	// 2) 输出用：写入 cgo 的相对路径（允许带 ..）
	emitRel := filepath.ToSlash(emitRelFromSRCDIR)

	var libs []string
	for _, name := range ordered {
		if !allSet[name] {
			fmt.Printf("Missing required static lib: %s (dir: %s)\n", name, baseDir)
			os.Exit(1)
		}
		libs = append(libs, "${SRCDIR}/"+filepath.ToSlash(filepath.Join(emitRel, name)))
		delete(allSet, name)
	}

	if len(allSet) > 0 {
		var extra []string
		for name := range allSet {
			extra = append(extra, name)
		}
		sort.Strings(extra)
		for _, name := range extra {
			libs = append(libs, "${SRCDIR}/"+filepath.ToSlash(filepath.Join(emitRel, name)))
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

func mustChmodX(path string) {
	st, err := os.Stat(path)
	if err != nil {
		fmt.Printf("Missing required file: %s, %v\n", path, err)
		os.Exit(1)
	}
	// 仅对非 windows 尝试 chmod +x
	if runtime.GOOS != "windows" {
		_ = os.Chmod(path, st.Mode().Perm()|0111)
	}
}

func fltkConfigPath(ctx *buildCtx) string {
	// 关键点：每个 arch 都用自己的 cmake build 目录下的 fltk-config
	// 这样 arm64 / amd64 不会串
	return filepath.Join(ctx.cmakeBuild, "bin", "fltk-config")
}

func runFltkConfig(ctx *buildCtx, args ...string) string {
	cfg := fltkConfigPath(ctx)
	mustChmodX(cfg)

	// fltk-config 是个 shell 脚本：windows 下也优先走 sh（MinGW/Git-Bash 都有）
	if ctx.goos == "windows" {
		out := outputCmd("", "sh", append([]string{cfg}, args...)...)
		return string(out)
	}

	out := outputCmd("", cfg, args...)
	return string(out)
}

func rewritePathsForCgo(ctx *buildCtx, s string) string {
	// cgo 文件在 fltk_bridge/ 下，项目根是 ${SRCDIR}/..
	// 统一把绝对 prefix 替换到 ${SRCDIR}/..
	//（你现在 install 到 outputRoot，然后又 copy 到项目根 lib/，所以要指向项目根）
	root := "${SRCDIR}/.."

	// fltk-config 里常出现 prefix=/abs/path/build/_install
	s = strings.ReplaceAll(s, ctx.outputRoot, root)
	s = strings.ReplaceAll(s, ctx.currentDir, root)

	// 保险：把多余换行/空格规范一下
	return strings.TrimSpace(s)
}

func cleanCMakeBuildDir(ctx *buildCtx) {
	_ = os.RemoveAll(ctx.cmakeBuild)
}

func buildStaticLibs(ctx *buildCtx) {
	prepareDirs(ctx)

	ensureFltkSource(ctx)
	checkoutTargetVersion(ctx)
	applyWindowsPatchIfNeeded(ctx)

	cleanCMakeBuildDir(ctx)
	runCMakeConfigure(ctx)
	runCMakeBuild(ctx)
	cleanStagingInstall(ctx)
	runCMakeInstall(ctx)

	moveFlConfigHeader(ctx)
	syncArtifactsToProjectRoot(ctx)

	writeManifestForTarget(ctx, ctx.goos, ctx.outArch)

	mustFileExist(filepath.Join("lib", ctx.goos, ctx.outArch, "FL", "fl_config.h"))
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
		writeManifestForTarget(ctxUni, "darwin", "universal")
		generateCgo(ctxUni)

		fmt.Println("Successfully built darwin arm64 + amd64 + universal")
		return
	}

	// ===== 其他平台 / 单架构 =====
	ctx := newBuildCtx(runtime.GOARCH, runtime.GOARCH)
	buildStaticLibs(ctx)
	generateCgo(ctx)

	fmt.Printf("Successfully generated libraries for OS: %s, architecture: %s\n",
		ctx.goos, ctx.goarch)
}
