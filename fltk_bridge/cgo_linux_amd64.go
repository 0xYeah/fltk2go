//go:build linux && amd64

package fltk_bridge

// #cgo linux,amd64 CPPFLAGS: -I${SRCDIR}/../lib/linux/amd64 -I${SRCDIR}/../lib/include -I${SRCDIR}/../lib/include/FL/images -D_LARGEFILE_SOURCE -D_LARGEFILE64_SOURCE -D_FILE_OFFSET_BITS=64
// #cgo linux,amd64 CXXFLAGS: -std=c++11
