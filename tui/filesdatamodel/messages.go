package filesdatamodel

import (
	"github.com/charmbracelet/bubbles/textinput"
)

type filePickerMsg struct {
	isDirPicker bool
}

type choosePathMsg struct {
	path string
}

type textInputMsg struct {
	inputs []textinput.Model
}

type closeTextInputMsg struct {
	value string
}
