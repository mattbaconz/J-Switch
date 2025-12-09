package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/user/jswitch/pkg/models"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")). // Pinkish
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")) // Greyish

	activeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")) // Green
)

type Model struct {
	installations []models.JavaInstallation
	cursor        int
	activeID      string
	quitting      bool
	SelectedID    string // Public field to retrieve selection after Run
}

func NewModel(installations []models.JavaInstallation, activeID string) Model {
	return Model{
		installations: installations,
		activeID:      activeID,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			} else {
				m.cursor = len(m.installations) - 1
			}
		case "down", "j":
			if m.cursor < len(m.installations)-1 {
				m.cursor++
			} else {
				m.cursor = 0
			}
		case "enter":
			if len(m.installations) > 0 {
				m.SelectedID = m.installations[m.cursor].Version
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	if m.quitting {
		return "Bye!\n"
	}
	if len(m.installations) == 0 {
		return "No installations found.\nPress q to quit.\n"
	}

	s := titleStyle.Render("J-Switch: Java Version Manager") + "\n\n"

	for i, inst := range m.installations {
		cursor := "  " // 2 spaces
		if m.cursor == i {
			cursor = "> "
		}

		checked := " "
		if inst.Version == m.activeID {
			checked = "✓"
		}

		// Row string: "[✓] Vendor Version (Path)"
		row := fmt.Sprintf("[%s] %-10s %-10s (%s)", checked, inst.Vendor, inst.Version, inst.Path)

		if m.cursor == i {
			s += selectedStyle.Render(cursor + row)
		} else if inst.Version == m.activeID {
			s += activeStyle.Render(cursor + row)
		} else {
			s += normalStyle.Render(cursor + row)
		}
		s += "\n"
	}

	s += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("↑/↓: Navigate • Enter: Switch • q: Quit") + "\n"
	return s
}
