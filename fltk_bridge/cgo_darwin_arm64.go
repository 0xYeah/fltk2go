//go:build darwin && arm64

package fltk_bridge

// #cgo darwin,arm64 CPPFLAGS: -I${SRCDIR}/../lib/darwin/arm64 -I${SRCDIR}/../lib/include -I${SRCDIR}/../lib/include/FL/images
// #cgo darwin,arm64 CXXFLAGS: -std=c++11
// #cgo darwin,arm64 LDFLAGS: ${SRCDIR}/../lib/darwin/arm64/libfltk_images.a ${SRCDIR}/../lib/darwin/arm64/libfltk_jpeg.a ${SRCDIR}/../lib/darwin/arm64/libfltk_png.a ${SRCDIR}/../lib/darwin/arm64/libfltk_z.a ${SRCDIR}/../lib/darwin/arm64/libfltk_gl.a ${SRCDIR}/../lib/darwin/arm64/libfltk_forms.a ${SRCDIR}/../lib/darwin/arm64/libfltk.a -lm -lpthread -framework Cocoa -framework OpenGL -framework UniformTypeIdentifiers
import "C"
