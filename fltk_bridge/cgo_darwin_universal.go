//go:build darwin && (amd64 || arm64)

package fltk_bridge

// #cgo darwin,arm64 CPPFLAGS: -I${SRCDIR}/../lib/darwin/universal -I${SRCDIR}/../lib/include -I${SRCDIR}/../lib/include/FL/images
// #cgo darwin,arm64 CXXFLAGS: -std=c++11
// #cgo darwin,arm64 LDFLAGS: ${SRCDIR}/../lib/darwin/universal/libfltk_images.a ${SRCDIR}/../lib/darwin/universal/libfltk_jpeg.a ${SRCDIR}/../lib/darwin/universal/libfltk_png.a ${SRCDIR}/../lib/darwin/universal/libfltk_z.a ${SRCDIR}/../lib/darwin/universal/libfltk_gl.a ${SRCDIR}/../lib/darwin/universal/libfltk_forms.a ${SRCDIR}/../lib/darwin/universal/libfltk.a -lm -lpthread -framework Cocoa -framework OpenGL -framework UniformTypeIdentifiers
import "C"
