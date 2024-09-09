// Package tui provides a text-based user interface for the Gophkeeper application.
// It includes functions to initialize and run the TUI program, handle user input,
// and display information to the user.
package tui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Tomap-Tomap/GophKeeper/tui/buildinfo"
	tea "github.com/charmbracelet/bubbletea"
)

const applicationFolder = ".gophkeeper"

// Run initializes and runs the TUI program with the provided context and build information.
func Run(ctx context.Context, buildInfo buildinfo.BuildInfo) error {
	dir, err := createApplicationFolder()

	if err != nil {
		return err
	}

	model := NewMainModel(ctx, buildInfo, dir)
	program := tea.NewProgram(model, tea.WithContext(ctx))

	_, err = program.Run()

	return err
}

func createApplicationFolder() (string, error) {
	homeDir, err := os.UserHomeDir()

	if err != nil {
		return "", fmt.Errorf("cannot get home dir: %w", err)
	}

	dirPath := filepath.Join(homeDir, applicationFolder)

	_, err = os.ReadDir(dirPath)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err := os.Mkdir(dirPath, os.ModeDir)

			if err != nil {
				return "", fmt.Errorf("cannot create application dir: %w", err)
			}
		} else {
			return "", fmt.Errorf("cannot read dir: %w", err)
		}
	}

	return dirPath, nil
}
