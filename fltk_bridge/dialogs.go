package fltk_bridge

/*
#include <stdlib.h>
#include "dialogs.h"
*/
import "C"
import "unsafe"

func MessageBox(title, message string) {
	titleStr := C.CString(title)
	defer C.free(unsafe.Pointer(titleStr))
	messageStr := C.CString(message)
	defer C.free(unsafe.Pointer(messageStr))
	C.go_fltk_message_box(titleStr, messageStr)
}

func ChoiceDialog(message string, options ...string) int {
	return choiceDialog("", message, options...)
}

// TitledChoiceDialog presents a native modal choice with an explicit window
// title. Security and destructive confirmations should prefer this over the
// legacy title-less ChoiceDialog so window managers and assistive tools expose
// the purpose of the prompt.
func TitledChoiceDialog(title, message string, options ...string) int {
	return choiceDialog(title, message, options...)
}

func choiceDialog(title, message string, options ...string) int {
	if len(options) == 0 || len(options) >= 3 {
		panic("Unsupported number of ChoiceDialog options")
	}
	messageStr := C.CString(message)
	defer C.free(unsafe.Pointer(messageStr))
	option0 := C.CString(options[0])
	defer C.free(unsafe.Pointer(option0))
	option1 := (*C.char)(nil)
	if len(options) > 1 {
		option1 = C.CString(options[1])
		defer C.free(unsafe.Pointer(option1))
	}
	option2 := (*C.char)(nil)
	if len(options) > 2 {
		option2 = C.CString(options[2])
		defer C.free(unsafe.Pointer(option2))
	}
	if title == "" {
		return int(C.go_fltk_choice_dialog(messageStr, option0, option1, option2))
	}
	titleStr := C.CString(title)
	defer C.free(unsafe.Pointer(titleStr))
	return int(C.go_fltk_titled_choice_dialog(titleStr, messageStr, option0, option1, option2))
}
