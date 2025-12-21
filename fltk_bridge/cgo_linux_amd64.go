//go:build linux && amd64

package fltk_bridge

// #cgo linux,amd64 CPPFLAGS: -I${SRCDIR}/../libs/fltk/include -I${SRCDIR}/../libs/fltk/linux/amd64 -I${SRCDIR}/../libs/fltk/include/FL/images -I${SRCDIR}/../libs/fltk/include -I${SRCDIR}/../libs/fltk/include/FL/images -D_LARGEFILE_SOURCE -D_LARGEFILE64_SOURCE -D_FILE_OFFSET_BITS=64 -D_THREAD_SAFE -D_REENTRANT
// #cgo linux,amd64 CXXFLAGS: -std=c++11
// #cgo linux,amd64 LDFLAGS: ${SRCDIR}/../libs/fltk/linux/amd64/libfltk_images.a ${SRCDIR}/../libs/fltk/linux/amd64/libfltk_jpeg.a ${SRCDIR}/../libs/fltk/linux/amd64/libfltk_png.a ${SRCDIR}/../libs/fltk/linux/amd64/libfltk_z.a ${SRCDIR}/../libs/fltk/linux/amd64/libfltk_gl.a -lGLU -lGL ${SRCDIR}/../libs/fltk/linux/amd64/libfltk_forms.a ${SRCDIR}/../libs/fltk/linux/amd64/libfltk.a -lm -lX11 -lXext -lpthread -lXinerama -lXfixes -lXcursor -lXft -lXrender -lfontconfig -ldl
import "C"
