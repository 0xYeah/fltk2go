
<!-- TOC -->

- [1. info](#1-info)
- [2. build FLTK](#2-build-fltk)
- [3. Use](#3-use)
- [4. Dir Tree](#4-dir-tree)
- [5. Warning](#5-warning)
- [6. used tools](#6-used-tools)
  - [6.1. tree](#61-tree)
    - [6.1.1. list projects](#611-list-projects)
  - [Depends](#depends)
    - [Linux](#linux)

<!-- /TOC -->

# 1. info
*  Use FLTK like the Cocoa Framework for iOS/Mac.
*  中文：像iOS/Mac的Cocoa Framework一样用FLTK
* [FLTK Resource Doc](https://www.fltk.org/doc-1.4/index.html)



# 2. build FLTK
```shell
go run fltk_build/fltk_build.go fltk_build/manifest.go
```

# 3. Use
```shell
go get github.com/0xYeah/fltk2go@latest
```

# 4. Dir Tree
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

# 5. Warning
*   Prevent GUI goroutines from being scheduled to other OS threads
*  中文：防止 GUI的goroutine 被调度到其他 OS 线程
```go
func main() {
    // 将当前 goroutine 绑定到当前操作系统线程。
    // 对于 Win32 / OpenGL / GDI+ 等具有线程亲和性的系统或 C/C++ API，这是必须的。
    // 防止 goroutine 被调度到其他 OS 线程，导致 GUI / 图形上下文失效或异常。
    runtime.LockOSThread()
	... // 其他代码
	
	
}
```

# 6. used tools
## 6.1. tree
### 6.1.1. list projects
```shell
tree -I ".git|build|lib"
```

## Depends
### Linux
```
apt update
apt install -y \
  build-essential cmake pkg-config \
  libx11-dev libxext-dev libxinerama-dev libxcursor-dev libxrender-dev libxfixes-dev \
  libxft-dev \
  libgl1-mesa-dev libglu1-mesa-dev \
  mesa-common-dev

```