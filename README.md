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