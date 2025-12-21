//go:build windows && amd64

package fltk_bridge

// #cgo windows,amd64 CPPFLAGS: -I${SRCDIR}/../libs/fltk/windows/amd64 -I${SRCDIR}/../libs/fltk/include -I${SRCDIR}/../libs/fltk/include/FL/images -D_LARGEFILE_SOURCE -D_LARGEFILE64_SOURCE -D_FILE_OFFSET_BITS=64
// #cgo windows,amd64 CXXFLAGS: -std=c++11
// #cgo windows,amd64 LDFLAGS: -mwindows ${SRCDIR}/../libs/fltk/windows/amd64/libfltk_images.a ${SRCDIR}/../libs/fltk/windows/amd64/libfltk_jpeg.a ${SRCDIR}/../libs/fltk/windows/amd64/libfltk_png.a ${SRCDIR}/../libs/fltk/windows/amd64/libfltk_z.a ${SRCDIR}/../libs/fltk/windows/amd64/libfltk_gl.a -lglu32 -lopengl32 ${SRCDIR}/../libs/fltk/windows/amd64/libfltk_forms.a ${SRCDIR}/../libs/fltk/windows/amd64/libfltk.a -lgdiplus -lole32 -luuid -lcomctl32 -lws2_32 -lwinspool -lfontconfig
import "C"
