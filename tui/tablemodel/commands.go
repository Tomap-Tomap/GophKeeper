package tablemodel

import (
	tea "github.com/charmbracelet/bubbletea"
)

func inputCmd() tea.Cmd {
	return func() tea.Msg {
		return inputMsg{}
	}
}

func updateCmd() tea.Cmd {
	return func() tea.Msg {
		return updateMsg{}
	}
}

func openCmd() tea.Cmd {
	return func() tea.Msg {
		return openMsg{}
	}
}

func enterCmd() tea.Cmd {
	return func() tea.Msg {
		return enterMsg{}
	}
}

func deleteCmd() tea.Cmd {
	return func() tea.Msg {
		return deleteMsg{}
	}
}
