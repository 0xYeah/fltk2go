# fltk2go
*  Use FLTK like the Cocoa Framework for iOS/Mac.
*  中文：像iOS/Mac的Cocoa Framework一样用FLTK
* [FLTK Resource Doc](https://www.fltk.org/doc-1.4/index.html)

# 列出项目目录
```shell
tree -I ".git|build|lib"
```

# build FLTK
```shell
go run fltk_build/fltk_build.go fltk_build/manifest.go
```

# Use
```shell
go get github.com/0xYeah/fltk2go@latest
```
# Tree
```shell
fltk2go/
├─ fltk2go.go          # Run / Quit / Version
├─ window/
├─ button/
├─ widget/
├─ runtime/            # ⭐ 运行时核心
│  ├─ handle/          # unsafe.Pointer 生命周期
│  ├─ callback/        # Go ↔ C 回调表
│  ├─ loop/            # UI loop / Run
│  └─ thread/          # 主线程约束
├─ fltk_bridge/        # C ABI / cgo
└─ lib/

```

# 着重注意
*  防止 GUI的goroutine 被调度到其他 OS 线程
```go
func main() {
    // 将当前 goroutine 绑定到当前操作系统线程。
    // 对于 Win32 / OpenGL / GDI+ 等具有线程亲和性的系统或 C/C++ API，这是必须的。
    // 防止 goroutine 被调度到其他 OS 线程，导致 GUI / 图形上下文失效或异常。
    runtime.LockOSThread()
	... // 其他代码
	
	
}
```

# 用到的工具
## changeLog Tool git-cliff
## Usage
```shell

```
### Windows-Install
```shell
winget install git-cliff
```
### Mac-Install
[Linux Binary releases Install](https://git-cliff.org/docs/installation/binary-releases)
### Linux-Install
```shell
winget install git-cliff
```
