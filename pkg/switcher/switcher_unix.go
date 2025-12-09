//go:build !windows

package switcher

import (
	"fmt"
	"os"
	"path/filepath"
)

func switchJava(javaPath string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	jswitchDir := filepath.Join(home, ".jswitch")
	linkPath := filepath.Join(jswitchDir, "current")

	// Remove existing link or file if it exists
	if _, err := os.Lstat(linkPath); err == nil {
		if err := os.Remove(linkPath); err != nil {
			return fmt.Errorf("failed to remove existing symlink: %w", err)
		}
	}

	// Create new symlink
	if err := os.Symlink(javaPath, linkPath); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	fmt.Printf("Success! Symlink updated at %s\n", linkPath)
	fmt.Println("Ensure this line is in your ~/.zshrc or ~/.bashrc:")
	fmt.Printf("export JAVA_HOME=%s\n", linkPath)
	fmt.Println("export PATH=$JAVA_HOME/bin:$PATH")

	return nil
}
