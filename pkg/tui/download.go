package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/user/jswitch/pkg/config"
	"github.com/user/jswitch/pkg/fetcher"
	"github.com/user/jswitch/pkg/models"
)

type progressMsg float64
type completionMsg string
type errMsg error

type DownloadModel struct {
	version      int
	progress     progress.Model
	percent      float64
	status       string
	done         bool
	err          error
	downloadUrl  string
	semver       string
	progressChan chan float64
}

func NewDownloadModel(version int) DownloadModel {
	return DownloadModel{
		version:  version,
		progress: progress.New(progress.WithDefaultGradient()),
		status:   fmt.Sprintf("Finding latest Java %d release...", version),
	}
}

func (m DownloadModel) Init() tea.Cmd {
	return findVersionCmd(m.version)
}

func findVersionCmd(version int) tea.Cmd {
	return func() tea.Msg {
		url, semver, err := fetcher.GetLatestVersion(version)
		if err != nil {
			return errMsg(err)
		}
		return foundVersionMsg{url: url, semver: semver}
	}
}

type foundVersionMsg struct {
	url    string
	semver string
}

func startDownloadCmd(url string, dest string, progChan chan float64) tea.Cmd {
	return func() tea.Msg {
		defer close(progChan)
		path, err := fetcher.DownloadAndExtract(url, dest, progChan)
		if err != nil {
			return errMsg(err)
		}
		return completionMsg(path)
	}
}

func listenForProgressCmd(sub chan float64) tea.Cmd {
	return func() tea.Msg {
		p, ok := <-sub
		if !ok {
			return nil
		}
		return progressMsg(p)
	}
}

func (m DownloadModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case foundVersionMsg:
		m.downloadUrl = msg.url
		m.semver = msg.semver
		m.status = fmt.Sprintf("Downloading Java %s...", msg.semver)

		// Determine destination
		home, _ := os.UserHomeDir()
		dest := filepath.Join(home, ".jswitch", "versions")
		os.MkdirAll(dest, 0755)

		m.progressChan = make(chan float64)

		return m, tea.Batch(
			startDownloadCmd(msg.url, dest, m.progressChan),
			listenForProgressCmd(m.progressChan),
		)

	case progressMsg:
		var cmds []tea.Cmd

		cmd := m.progress.SetPercent(float64(msg))
		cmds = append(cmds, cmd)

		m.percent = float64(msg)

		if msg < 1.0 {
			cmds = append(cmds, listenForProgressCmd(m.progressChan))
		}
		return m, tea.Batch(cmds...)

	case completionMsg:
		m.status = fmt.Sprintf("Installed successfully to: %s", msg)
		m.done = true
		m.percent = 1.0

		// Update Config
		cfg, err := config.LoadConfig()
		if err == nil {
			cfg.Installations = append(cfg.Installations, models.JavaInstallation{
				Vendor:  "Eclipse Adoptium",
				Version: m.semver,
				Path:    string(msg),
			})
			if err := config.SaveConfig(cfg); err != nil {
				m.status += fmt.Sprintf("\nFailed to save config: %v", err)
			} else {
				m.status += "\nConfig updated."
			}
		} else {
			m.status += fmt.Sprintf("\nFailed to load config: %v", err)
		}

		m.status += "\nPress q to quit."
		return m, nil

	case errMsg:
		m.err = msg
		m.status = fmt.Sprintf("Error: %v\nPress q to quit.", msg)
		return m, nil

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd
	}
	return m, nil
}

func (m DownloadModel) View() string {
	if m.err != nil {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(m.status) + "\n"
	}

	pad := "\n" + strings.Repeat(" ", 2)
	s := "\n" +
		pad + titleStyle.Render("J-Switch Downloader") + "\n" +
		pad + m.status + "\n" +
		pad + m.progress.View() + "\n"

	if m.done {
		s += pad + lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("Done!") + "\n"
	}
	return s
}
