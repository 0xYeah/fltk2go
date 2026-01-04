//go:build darwin && (amd64 || arm64)

package fltk_bridge

// #cgo darwin CPPFLAGS: -I${SRCDIR}/../libs/fltk/darwin/universal -I${SRCDIR}/../libs/fltk/include -I${SRCDIR}/../libs/fltk/include/FL/images -isysroot /Library/Developer/CommandLineTools/SDKs/MacOSX.sdk -D_LARGEFILE_SOURCE -D_LARGEFILE64_SOURCE -D_FILE_OFFSET_BITS=64 -D_THREAD_SAFE -D_REENTRANT
// #cgo darwin CXXFLAGS: -std=c++11
// #cgo darwin LDFLAGS: ${SRCDIR}/../libs/fltk/darwin/universal/libfltk_images.a ${SRCDIR}/../libs/fltk/darwin/universal/libfltk_jpeg.a ${SRCDIR}/../libs/fltk/darwin/universal/libfltk_png.a ${SRCDIR}/../libs/fltk/darwin/universal/libfltk_z.a ${SRCDIR}/../libs/fltk/darwin/universal/libfltk_gl.a -framework OpenGL ${SRCDIR}/../libs/fltk/darwin/universal/libfltk_forms.a ${SRCDIR}/../libs/fltk/darwin/universal/libfltk.a -lm -lpthread -framework Cocoa -framework UniformTypeIdentifiers
import "C"
