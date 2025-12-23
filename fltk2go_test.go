package fltk2go

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestSupportedVersion(t *testing.T) {
	as, _ := GetSupportedLibraries()

	b, _ := json.MarshalIndent(as, "", "  ")
	fmt.Println(string(b))
}
