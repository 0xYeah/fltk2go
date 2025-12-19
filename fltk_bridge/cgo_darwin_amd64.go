//go:build darwin && amd64

package fltk_bridge

// #cgo darwin,amd64 CPPFLAGS: -I${SRCDIR}/../lib/darwin/amd64 -I${SRCDIR}/../lib/include -I${SRCDIR}/../lib/include/FL/images
// #cgo darwin,amd64 CXXFLAGS: -std=c++11
// #cgo darwin,amd64 LDFLAGS: ${SRCDIR}/../lib/darwin/amd64/libfltk_images.a ${SRCDIR}/../lib/darwin/amd64/libfltk_jpeg.a ${SRCDIR}/../lib/darwin/amd64/libfltk_png.a ${SRCDIR}/../lib/darwin/amd64/libfltk_z.a ${SRCDIR}/../lib/darwin/amd64/libfltk_gl.a ${SRCDIR}/../lib/darwin/amd64/libfltk_forms.a ${SRCDIR}/../lib/darwin/amd64/libfltk.a -lm -lpthread -framework Cocoa -framework OpenGL -framework UniformTypeIdentifiers
import "C"
