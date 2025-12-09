package scanner

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/user/jswitch/pkg/models"
)

// Common paths to ignore to speed up scanning
var ignoredDirs = map[string]bool{
	"Windows":       true,
	"ProgramData":   true,
	"$Recycle.Bin":  true,
	"System Volume Information": true,
	".git":          true,
	"node_modules":  true,
}

// ScanSystem recursively scans the provided root paths for Java installations.
func ScanSystem(rootPaths []string) ([]models.JavaInstallation, error) {
	var installations []models.JavaInstallation
	seenPaths := make(map[string]bool)

	for _, root := range rootPaths {
        // Check if root exists first
        if _, err := os.Stat(root); os.IsNotExist(err) {
            continue
        }

		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				// Permission errors or path errors should not stop the entire scan
				return nil
			}

			if d.IsDir() {
				if ignoredDirs[d.Name()] {
					return filepath.SkipDir
				}
				// Potential check: if this directory looks like a JDK root (contains bin/java),
				// we might want to check it and then SkipDir to avoid scanning inside it?
				// For now, let's just let it walk.
				return nil
			}

			// We are looking for the java executable
			name := strings.ToLower(d.Name())
			if name == "java" || name == "java.exe" {
				// Check if it's inside a 'bin' directory
				dir := filepath.Dir(path)
				if strings.EqualFold(filepath.Base(dir), "bin") {
					// Found a candidate
					installPath := filepath.Dir(dir) // parent of bin is the install root
					
					// Avoid duplicates
					if seenPaths[installPath] {
						return nil
					}

					inst, err := verifyAndParseJava(path, installPath)
					if err == nil {
						installations = append(installations, inst)
						seenPaths[installPath] = true
					}
				}
			}

			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("error walking path %s: %w", root, err)
		}
	}

	return installations, nil
}

// verifyAndParseJava runs 'java -version' and parses the output.
func verifyAndParseJava(exePath, installRoot string) (models.JavaInstallation, error) {
	cmd := exec.Command(exePath, "-version")
	// java -version writes to stderr
	var out bytes.Buffer
	cmd.Stderr = &out
	// Some implementations might write to stdout
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return models.JavaInstallation{}, err
	}

	output := out.String()
	return parseVersionOutput(output, installRoot)
}

func parseVersionOutput(output, installRoot string) (models.JavaInstallation, error) {
	// Example output:
	// openjdk version "17.0.2" 2022-01-18
	// java version "1.8.0_202"

	// Regex to capture version string
	// look for version "..."
	versionRegex := regexp.MustCompile(`version "([^"]+)"`)
	matches := versionRegex.FindStringSubmatch(output)
	if len(matches) < 2 {
		return models.JavaInstallation{}, fmt.Errorf("could not parse version string")
	}

	versionStr := matches[1]
	
	// Determine vendor (heuristics)
	vendor := "Unknown"
	lowerOutput := strings.ToLower(output)
	if strings.Contains(lowerOutput, "openjdk") {
		vendor = "OpenJDK"
	} else if strings.Contains(lowerOutput, "java(tm)") || strings.Contains(lowerOutput, "hotspot") {
		vendor = "Oracle"
	} else if strings.Contains(lowerOutput, "zulu") {
		vendor = "Azul Zulu"
	}
	// Add more vendor checks as needed

	// Determine major version
	major := 0
	if strings.HasPrefix(versionStr, "1.") {
		// Old style: 1.8.0 -> 8
		parts := strings.Split(versionStr, ".")
		if len(parts) >= 2 {
			if val, err := strconv.Atoi(parts[1]); err == nil {
				major = val
			}
		}
	} else {
		// New style: 17.0.2 -> 17
		parts := strings.Split(versionStr, ".")
		if len(parts) > 0 {
			if val, err := strconv.Atoi(parts[0]); err == nil {
				major = val
			}
		}
	}

	return models.JavaInstallation{
		Version:      versionStr,
		MajorVersion: major,
		Path:         installRoot,
		Vendor:       vendor,
	}, nil
}
