package dialog

import "github.com/0xdevelop/fltk2go/fltk_bridge"

// Message presents an informational modal dialog.
func Message(title, message string) {
	fltk_bridge.MessageBox(title, message)
}

// Alert is an alias for Message for UIKit-style call sites.
func Alert(title, message string) {
	Message(title, message)
}

// Choice presents a modal choice dialog and returns the selected option index.
// fltk_bridge currently supports one or two options; invalid option counts are
// converted to a safe "cancel" result instead of panicking at the UIKit layer.
func Choice(message string, options ...string) int {
	if len(options) == 0 || len(options) > 2 {
		return -1
	}
	return fltk_bridge.ChoiceDialog(message, options...)
}

// TitledChoice presents a modal choice dialog with an explicit native window
// title. Use it when the title carries important security or destructive-action
// context that must remain visible to the window manager and accessibility UI.
func TitledChoice(title, message string, options ...string) int {
	if len(options) == 0 || len(options) > 2 {
		return -1
	}
	return fltk_bridge.TitledChoiceDialog(title, message, options...)
}
