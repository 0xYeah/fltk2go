package fltk2go

import (
	"fmt"
	"testing"
)

func TestSupportedVersion(t *testing.T) {
	as, _ := GetSupportedLibraries()
	fmt.Println(as)
}
