package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/Tomap-Tomap/GophKeeper/client"
	"github.com/Tomap-Tomap/GophKeeper/crypto"
	"github.com/Tomap-Tomap/GophKeeper/tui/buildinfo"
	"github.com/Tomap-Tomap/GophKeeper/tui/colors"
	"github.com/Tomap-Tomap/GophKeeper/tui/columns"
	"github.com/Tomap-Tomap/GophKeeper/tui/commands"
	"github.com/Tomap-Tomap/GophKeeper/tui/config"
	"github.com/Tomap-Tomap/GophKeeper/tui/configmodel"
	"github.com/Tomap-Tomap/GophKeeper/tui/constants"
	"github.com/Tomap-Tomap/GophKeeper/tui/messages"
	"github.com/Tomap-Tomap/GophKeeper/tui/startmodel"
	"github.com/Tomap-Tomap/GophKeeper/tui/tablemodel"
	"github.com/Tomap-Tomap/GophKeeper/tui/tabsmodel"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var helpStyle = lipgloss.NewStyle().Foreground(colors.HelpColor)

// Messager is an interface for getting the model
type Messager interface {
	GetModel() (Model, tea.Cmd, string)
}

// Model represents the main model for the TUI application
type Model struct {
	ctx context.Context

	terminalBorder lipgloss.Style
	infoBorder     lipgloss.Style

	currentModel tea.Model

	config *config.Config

	infoText string
	helpText string
}

// NewMainModel creates a new main model for the TUI application
func NewMainModel(ctx context.Context, buildInfo buildinfo.BuildInfo, dir string) Model {
	welcomeMessage := strings.Builder{}
	welcomeMessage.WriteString("Welcome to Goph Keeper! ")

	config, err := config.New(dir)

	if err != nil {
		welcomeMessage.WriteString(err.Error())
		welcomeMessage.WriteString(" ")
	}

	welcomeMessage.WriteString(buildInfo.String())

	border := lipgloss.NormalBorder()

	tb := lipgloss.NewStyle().
		Border(border).
		BorderForeground(colors.MainColor).Align(lipgloss.Center, lipgloss.Bottom)

	ib := lipgloss.NewStyle().
		Border(border).
		BorderForeground(colors.MainColor)

	return Model{
		ctx:            ctx,
		terminalBorder: tb,
		infoBorder:     ib,
		currentModel:   startmodel.New(),
		config:         config,
		infoText:       welcomeMessage.String(),
		helpText:       "↑: move up • ↓: move down • enter: select",
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update updates the model based on the received message
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if km, ok := msg.(tea.KeyMsg); ok && km.Type == tea.KeyCtrlC {
		return m, tea.Quit
	}

	if wm, ok := msg.(tea.WindowSizeMsg); ok {
		m, msg = m.calculateSize(wm)
	}

	switch msg := msg.(type) {
	case messages.OpenConfigModel:
		m.currentModel = configmodel.New(m.config)

		return m, tea.Batch(commands.SetWindowSize(), m.currentModel.Init())
	case messages.CloseConfigModel:
		m.config.AddrToService = msg.AddrToService
		m.config.PathToSecretKey = msg.PathToKey

		err := m.config.Save()

		if err != nil {
			return m, commands.Error(err)
		}

		m.currentModel = startmodel.New()

		return m, tea.Batch(commands.SetWindowSize(), m.currentModel.Init())
	case messages.SignIn:
		client, err := newGrpcClient(m.config)

		if err != nil {
			return m, commands.Error(err)
		}

		err = client.SignIn(m.ctx, msg.Login, msg.Password)

		if err != nil {
			return m, commands.Error(err)
		}

		currentModel, err := tabsmodel.New([]tablemodel.Columner{
			columns.NewPasswordColumns(m.ctx, client),
			columns.NewBanksColumns(m.ctx, client),
			columns.NewTextColumns(m.ctx, client),
			columns.NewFileColumns(m.ctx, client),
		}, []string{"Passwords", "Banks", "Texts", "Files"})

		if err != nil {
			return m, commands.Error(err)
		}

		m.currentModel = currentModel
		return m, tea.Batch(commands.SetWindowSize(), m.currentModel.Init())
	case messages.Registration:
		client, err := newGrpcClient(m.config)

		if err != nil {
			return m, commands.Error(err)
		}

		err = client.Register(m.ctx, msg.Login, msg.Password)

		if err != nil {
			return m, commands.Error(err)
		}

		currentModel, err := tabsmodel.New([]tablemodel.Columner{
			columns.NewPasswordColumns(m.ctx, client),
			columns.NewBanksColumns(m.ctx, client),
			columns.NewTextColumns(m.ctx, client),
			columns.NewFileColumns(m.ctx, client),
		}, []string{"Passwords", "Banks", "Texts", "Files"})

		if err != nil {
			return m, commands.Error(err)
		}

		m.currentModel = currentModel
		return m, tea.Batch(commands.SetWindowSize(), m.currentModel.Init())
	case messages.Info:
		m.infoText = msg.Info
		m.helpText = msg.Help
	case messages.Error:
		m.infoText = msg.Err.Error()
	}

	m.currentModel, cmd = m.currentModel.Update(msg)

	return m, cmd
}

// View returns the view of the model as a string
func (m Model) View() string {
	infoBorder := m.infoBorder.Render(m.infoText + " " + helpStyle.Render(m.helpText))
	modelView := lipgloss.JoinVertical(lipgloss.Center, m.currentModel.View(), infoBorder)

	return m.terminalBorder.Render(modelView)
}

func (m Model) calculateSize(msg tea.WindowSizeMsg) (Model, tea.WindowSizeMsg) {
	newWidth := msg.Width - m.terminalBorder.GetHorizontalFrameSize()
	newHeight := msg.Height - m.terminalBorder.GetVerticalFrameSize()

	m.terminalBorder = m.terminalBorder.
		Width(newWidth).
		Height(newHeight)

	newIH := newHeight/10 - m.infoBorder.GetVerticalFrameSize()

	if newIH <= 0 {
		newIH = constants.StringHeight
	}

	newIHWB := newIH + m.infoBorder.GetVerticalFrameSize()

	m.infoBorder = m.infoBorder.
		Width(newWidth - m.infoBorder.GetHorizontalFrameSize()).
		Height(newIH)

	return m, tea.WindowSizeMsg{
		Width:  newWidth,
		Height: newHeight - newIHWB,
	}
}

func newGrpcClient(config *config.Config) (*client.Client, error) {
	crypter, err := crypto.NewCrypterByFile(config.PathToSecretKey)

	if err != nil {
		return nil, fmt.Errorf("cannot create crypter: %w", err)
	}

	client, err := client.New(crypter, config.AddrToService)

	if err != nil {
		return nil, fmt.Errorf("cannot create grpc client: %w", err)
	}

	return client, nil
}
