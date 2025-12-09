package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"text/tabwriter"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/user/jswitch/pkg/config"
	"github.com/user/jswitch/pkg/scanner"
	"github.com/user/jswitch/pkg/switcher"
	"github.com/user/jswitch/pkg/tui"
)

func main() {
	if len(os.Args) < 2 {
		// Default to UI if no args provided (friendly for double-clicking)
		handleUI()
		return
	}

	command := os.Args[1]

	switch command {
	case "scan":
		// Allow passing custom paths after "scan"
		customPaths := os.Args[2:]
		handleScan(customPaths)
	case "list":
		handleList()
	case "use":
		if len(os.Args) < 3 {
			fmt.Println("Usage: jswitch use <version>")
			return
		}
		version := os.Args[2]
		handleUse(version)
	case "ui", "select":
		handleUI()
	case "install":
		if len(os.Args) < 3 {
			fmt.Println("Usage: jswitch install <version>")
			return
		}
		versionStr := os.Args[2]
		version, err := strconv.Atoi(versionStr)
		if err != nil {
			fmt.Println("Version must be an integer (e.g. 17)")
			return
		}
		handleInstall(version)
	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Usage: jswitch <command> [arguments]")
	fmt.Println("Resulting Binary: jswitch")
	fmt.Println("\nCommands:")
	fmt.Println("  ui                Open interactive selection menu")
	fmt.Println("  scan [paths...]   Scan system for Java installations")
	fmt.Println("  list              List discovered Java versions")
	fmt.Println("  use <version>     Select a Java version to use")
	fmt.Println("  install <version> Download and install a Java version (e.g. 17)")
}

func handleScan(customPaths []string) {
	var pathsToScan []string
	if len(customPaths) > 0 {
		pathsToScan = customPaths
	} else {
		if runtime.GOOS == "windows" {
			pathsToScan = []string{
				`C:\Program Files\Java`,
				`C:\Program Files (x86)\Java`,
			}
		} else {
			pathsToScan = []string{
				"/usr/lib/jvm",
				"/usr/java",
				"/Library/Java/JavaVirtualMachines",
			}
		}
	}

	fmt.Printf("Scanning paths: %v\n", pathsToScan)

	installations, err := scanner.ScanSystem(pathsToScan)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning: %v\n", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		// If load fails, start fresh
		cfg = &config.Config{}
	}

	if len(installations) > 0 {
		fmt.Printf("Found %d Java installations.\n", len(installations))
		cfg.Installations = installations
		if err := config.SaveConfig(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		} else {
			home, _ := os.UserHomeDir()
			fmt.Printf("Config saved to %s\n", filepath.Join(home, ".jswitch", "config.json"))
		}
	} else {
		fmt.Println("No Java installations found.")
	}
}

func handleList() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		return
	}

	if len(cfg.Installations) == 0 {
		fmt.Println("No installations found. Run 'jswitch scan' first.")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "CURRENT\tVENDOR\tVERSION\tPATH")

	for _, inst := range cfg.Installations {
		marker := " "
		if inst.Version == cfg.CurrentVersion {
			marker = "*"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", marker, inst.Vendor, inst.Version, inst.Path)
	}
	w.Flush()
}

func handleUse(version string) {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		return
	}

	found := false
	for _, inst := range cfg.Installations {
		if inst.Version == version {
			cfg.CurrentVersion = inst.Version
			found = true
			break
		}
	}

	if !found {
		fmt.Printf("Version %s not found. Run 'jswitch list' to see options.\n", version)
		return
	}

	if err := config.SaveConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		return
	}

	fmt.Printf("Target set to Java %s.\n", version)

	// Apply system changes
	if err := switcher.Switch(cfg.CurrentVersionPath(version)); err != nil {
		fmt.Fprintf(os.Stderr, "Error switching system environment: %v\n", err)
	}
}

func handleUI() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		return
	}

	if len(cfg.Installations) == 0 {
		fmt.Println("No installations found. Run 'jswitch scan' first.")
		return
	}

	initialModel := tui.NewModel(cfg.Installations, cfg.CurrentVersion)
	p := tea.NewProgram(initialModel)

	// Run the program
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		return
	}

	// Assert back to our specific model
	m, ok := finalModel.(tui.Model)
	if ok && m.SelectedID != "" {
		// Reuse handleUse logic to apply switch
		fmt.Printf("Selected via UI: %s\n", m.SelectedID)
		handleUse(m.SelectedID)
	}
}

func handleInstall(version int) {
	m := tui.NewDownloadModel(version)
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running installer: %v\n", err)
	}
}
