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

	// staging install 根目录（给 CMake install 用；不要求最终产物落在这）
	// 默认: build/_install
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

	// ✅ 最终产物固定落到这里（全平台统一）
	FinalRoot = getEnvOrDefault("FLTK_FINAL_ROOT", filepath.Join("libs", "fltk"))
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

	// staging install 输出相关（CMake install 写这里）
	outputRoot string
	libdir     string
	includeDir string

	// ✅ 最终产物根目录（copy 到这里）
	finalRoot string

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

	// staging install: build/_install/lib/<os>/<arch>
	libdir := filepath.Join(outputRoot, "lib", goos, outArch)
	includeDir := filepath.Join(outputRoot, "include")

	finalRoot := filepath.Clean(FinalRoot)

	return &buildCtx{
		goos:       goos,
		goarch:     goarch,
		outArch:    outArch,
		outputRoot: outputRoot,
		libdir:     libdir,
		includeDir: includeDir,
		finalRoot:  finalRoot,

		buildRoot:  buildRoot,
		fltkSource: filepath.Join(buildRoot, "fltk"),
		// 必须按 arch 隔离，否则一定踩坑
		cmakeBuild: filepath.Join(buildRoot, fmt.Sprintf("fltk-cmake-%s-%s", goos, outArch)),

		currentDir: mustGetwd(),
	}
}

func prepareDirs(ctx *buildCtx) {
	mustMkdirAll(ctx.buildRoot, 0750)
	mustMkdirAll(ctx.libdir, 0750)
	mustMkdirAll(ctx.includeDir, 0750)

	// ✅ 最终根目录也预建
	mustMkdirAll(ctx.finalRoot, 0750)
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
		"-DFLTK_BUILD_SHARED_LIBS=OFF",

		// bundled libs (跨发行版/跨架构稳定)
		"-DFLTK_USE_SYSTEM_LIBJPEG=OFF",
		"-DFLTK_USE_SYSTEM_LIBPNG=OFF",
		"-DFLTK_USE_SYSTEM_ZLIB=OFF",

		// OpenGL
		"-DFLTK_USE_GL=ON",

		// Wayland（你当前策略：禁用，避免依赖漂移）
		"-DFLTK_USE_WAYLAND=OFF",

		// install 到 staging outputRoot
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

// ✅ 保持你原来的策略：把 fl_config.h 从 include/FL 移到 lib/<os>/<arch>/FL
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
	// 统一空白，避免 "  " 影响 contains 判定
	s = " " + strings.TrimSpace(s) + " "

	for _, f := range flags {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		// 用两侧空格做粗略“单词边界”，比 strings.Contains 更稳一点
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

	// 文件名：darwin universal 特例
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

	// universal 的 #cgo 条件必须用 darwin（不带 arch）
	cgoCondOSArch := fmt.Sprintf("%s,%s", ctx.goos, ctx.goarch)
	if isDarwinUniversal {
		cgoCondOSArch = "darwin"
	}

	// 1) 通过 fltk-config 获取参数
	cxx := strings.TrimSpace(runFltkConfig(ctx, "--use-gl", "--use-images", "--use-forms", "--cxxflags"))
	ld := strings.TrimSpace(runFltkConfig(ctx, "--use-gl", "--use-images", "--use-forms", "--ldstaticflags"))

	// 2) rewrite 到最终目录（libs/fltk）
	cxx = strings.TrimSpace(rewritePathsForCgo(ctx, cxx))
	ld = strings.TrimSpace(rewritePathsForCgo(ctx, ld))

	// 3) macOS：weak_framework 兼容处理（建议统一成 -framework，兼容性更好）
	if ctx.goos == "darwin" {
		ld = strings.ReplaceAll(ld, "-weak_framework UniformTypeIdentifiers", "-framework UniformTypeIdentifiers")
	}

	// 4) largefile：按“缺啥补啥”
	// （macOS 一般不需要你这三项；linux/windows 常见需要）
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

	// 5) 让 fl_config.h 的路径最优先：把 -I<final>/<os>/<arch> 放最前
	finalRoot := "${SRCDIR}/../" + filepath.ToSlash(ctx.finalRoot)
	incArch := fmt.Sprintf("-I%s/%s/%s", finalRoot, ctx.goos, ctx.outArch) // 最前（命中 FL/fl_config.h）
	incBase := fmt.Sprintf("-I%s/include", finalRoot)
	incImg := fmt.Sprintf("-I%s/include/FL/images", finalRoot)

	// 关键修复：不要 ReplaceAll 删子串，改为 token 级过滤 & 去重
	rmSet := map[string]struct{}{
		incArch: {},
		incBase: {},
		incImg:  {},
	}

	toks := strings.Fields(cxx)
	kept := make([]string, 0, len(toks))
	seen := make(map[string]struct{}, len(toks))

	for _, t := range toks {
		// 过滤掉我们自己会前置补上的 include 路径（避免重复）
		if _, ok := rmSet[t]; ok {
			continue
		}

		// 保险：如果之前错误 ReplaceAll 产生过裸 "/FL/images"，这里直接丢弃
		//（防止 invalid flag in #cgo CPPFLAGS: /FL/images）
		if strings.HasPrefix(t, "/") && strings.Contains(t, "FL/images") && !strings.HasPrefix(t, "-I") {
			continue
		}

		// token 去重（完全相同的 flag 不要重复）
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		kept = append(kept, t)
	}
	cxx = strings.Join(kept, " ")

	// 6) 拼出最终 CPPFLAGS：我们自己的 -I 永远在最前
	// 你要求 fl_config.h 相关优先级最高，所以顺序：incArch -> incBase -> incImg -> cxx(剩余 flags)
	cppflags := strings.TrimSpace(strings.Join([]string{incArch, incBase, incImg, cxx}, " "))

	// Windows：保持 -mwindows（也去重）
	if ctx.goos == "windows" {
		ld = removeTokens(ld, "-lfontconfig")

		if !strings.Contains(ld, "-mwindows") {
			ld = "-mwindows " + ld
		}
	}

	// 输出 #cgo
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

// ✅ 最终产物同步：统一落到 libs/fltk（全平台一致）
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
	return filepath.Join(ctx.cmakeBuild, "bin", "fltk-config")
}

func findGitShPath() (string, error) {
	// 常见的Git安装路径
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

	// 也可以通过环境变量GIT_HOME查找
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
		// 自动查找Git的sh.exe路径
		shPath, err := findGitShPath()
		if err != nil {
			fmt.Printf("Failed to find sh.exe: %v, fallback to 'sh'\n", err)
			shPath = "sh"
		}
		// 用绝对路径的sh.exe调用，解决Goland调试路径问题
		cmdArgs := append([]string{"-c", fmt.Sprintf("\"%s\" %s", cfg, strings.Join(args, " "))})
		out := outputCmd("", shPath, cmdArgs...)
		return string(out)
	}
	out := outputCmd("", cfg, args...)
	return string(out)
}

// ✅ 关键改动：rewrite 的 root 指向最终产物根 libs/fltk，而不是 ${SRCDIR}/..
func rewritePathsForCgo(ctx *buildCtx, s string) string {
	// cgo 文件在 fltk_bridge/ 下：${SRCDIR} = <repo>/fltk_bridge
	// 最终产物在 <repo>/libs/fltk
	root := "${SRCDIR}/../" + filepath.ToSlash(ctx.finalRoot)

	// 统一成 forward slash，避免 windows 的 \ 干扰 Replace
	ss := strings.ReplaceAll(s, "\\", "/")

	// 计算 staging 的可能路径形态（相对 + 绝对 + 关键子目录）
	outRel := filepath.ToSlash(filepath.Clean(ctx.outputRoot))

	outAbs, _ := filepath.Abs(ctx.outputRoot)
	outAbs = filepath.ToSlash(filepath.Clean(outAbs))

	incAbs, _ := filepath.Abs(filepath.Join(ctx.outputRoot, "include"))
	incAbs = filepath.ToSlash(filepath.Clean(incAbs))

	libAbs, _ := filepath.Abs(filepath.Join(ctx.outputRoot, "lib", ctx.goos, ctx.outArch))
	libAbs = filepath.ToSlash(filepath.Clean(libAbs))

	wdAbs := filepath.ToSlash(filepath.Clean(ctx.currentDir))

	// 依次替换（从更具体到更泛化）
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

	// 压一压空白
	return strings.TrimSpace(ss)
}

func cleanCMakeBuildDir(ctx *buildCtx) {
	_ = os.RemoveAll(ctx.cmakeBuild)
}

// =======================
// macOS universal merge (保持原逻辑，只改目录到 libs/fltk/...)
// =======================
func mergeDarwinUniversal(ctx *buildCtx) {
	if ctx.goos != "darwin" || ctx.outArch != "universal" {
		return
	}

	fmt.Println("Merging macOS universal binaries (lipo)")

	// ✅ 从最终目录合并（因为我们最终要落到 libs/fltk）
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

		runCmd("", "lipo", "-create", armLib, amdLib, "-output", outLib)
	}

	// fl_config.h：任选一个
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
// 主流程
// =======================
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

	// fl_config.h 从 include/FL -> staging <os>/<arch>/FL
	moveFlConfigHeader(ctx)

	// ✅ staging -> final: libs/fltk/include + libs/fltk/lib/<os>/<arch>
	syncArtifactsToFinalRoot(ctx)

	// ✅ 校验最终产物
	mustFileExist(filepath.Join(ctx.finalRoot, ctx.goos, ctx.outArch, "FL", "fl_config.h"))
	mustFileExist(filepath.Join(ctx.finalRoot, ctx.goos, ctx.outArch, "libfltk.a"))

	// 你原来的 manifest 写入保留（如果你项目里有这个函数）
	writeManifestForTarget(ctx, ctx.goos, ctx.outArch)
}

func main() {
	mustCheckEnv()
	mustCheckTool("git")
	mustCheckTool("cmake")

	if runtime.GOOS == "darwin" {
		// 先产出薄库到 libs/fltk/darwin/{arm64,amd64}
		buildStaticLibs(newBuildCtx("arm64", "arm64"))
		buildStaticLibs(newBuildCtx("amd64", "amd64"))

		// 再合并成 universal 到 libs/fltk/darwin/universal
		ctxUni := newBuildCtx("arm64", "universal")
		mergeDarwinUniversal(ctxUni)

		// manifest + cgo
		writeManifestForTarget(ctxUni, "darwin", "universal")
		generateCgo(ctxUni)

		fmt.Println("Successfully built darwin arm64 + amd64 + universal")
		return
	}

	// ===== 其他平台 / 单架构 =====
	ctx := newBuildCtx(runtime.GOARCH, runtime.GOARCH)
	buildStaticLibs(ctx)
	generateCgo(ctx)

	fmt.Printf("Successfully generated libraries for OS: %s, architecture: %s\n", ctx.goos, ctx.goarch)
}
