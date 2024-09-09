// Package colors provides a set of predefined colors using the lipgloss library.
// These colors can be used for styling terminal applications.
package colors

import "github.com/charmbracelet/lipgloss"

var (
	// MainColor is the primary color used for general styling.
	MainColor = lipgloss.Color("#755d9a")

	// FocusColor is used to highlight focused elements.
	FocusColor = lipgloss.Color("#ff3f18")

	// HelpColor is used for help text or secondary information.
	HelpColor = lipgloss.Color("241")
)
