
<!-- TOC -->

- [1. info](#1-info)
- [2. build FLTK](#2-build-fltk)
- [3. Use](#3-use)
- [4. Dir Tree](#4-dir-tree)
- [5. Warning](#5-warning)
- [6. used tools](#6-used-tools)
  - [6.1. tree](#61-tree)
    - [6.1.1. list projects](#611-list-projects)
  - [6.2. changeLog Tool git-cliff](#62-changelog-tool-git-cliff)
    - [6.2.1. Usage](#621-usage)
    - [6.2.2. Windows-Install](#622-windows-install)
    - [6.2.3. Mac-Install](#623-mac-install)
    - [6.2.4. Linux-Install](#624-linux-install)
  - [6.3. commit message style](#63-commit-message-style)
  - [6.4. use git-cliff example](#64-use-git-cliff-example)

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

## 6.2. changeLog Tool git-cliff
### 6.2.1. Usage
```shell
 git-cliff --config cliff.toml "v0.0.2..HEAD" --tag "v0.0.3" -o test_changelog.md
```

### 6.2.2. Windows-Install
```shell
winget install git-cliff
```
### 6.2.3. Mac-Install
[Linux Binary releases Install](https://git-cliff.org/docs/installation/binary-releases)
### 6.2.4. Linux-Install
```shell
winget install git-cliff
```

## 6.3. commit message style
| type       | 含义        | 是否进 changelog |
| ---------- | --------- | ------------- |
| `feat`     | 新功能       | ✅             |
| `fix`      | Bug 修复    | ✅             |
| `docs`     | 文档        | 可选            |
| `refactor` | 重构（无功能变化） | 可选            |
| `perf`     | 性能优化      | ✅             |
| `test`     | 测试        | ❌ / 可选        |
| `build`    | 构建系统      | ❌             |
| `ci`       | CI 配置     | ❌             |
| `chore`    | 杂项        | ❌             |
| `style`    | 格式（不影响逻辑） | ❌             |
| `revert`   | 回滚        | ✅             |

## 6.4. use git-cliff example
```shell
git commit -m "feat(fltk): add Win32 font patch"

git commit -m "fix(cgo): resolve link order on Windows"
```