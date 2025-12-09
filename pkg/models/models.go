package models

import "fmt"

// JavaInstallation represents a detected Java JDK/JRE on the system.
type JavaInstallation struct {
	// Version is the parsed version string (e.g., "17.0.2", "1.8.0_202").
	Version string `json:"version"`
	// MajorVersion is the primary version number (e.g., 8, 11, 17) for easy sorting/filtering.
	MajorVersion int `json:"major_version"`
	// Path is the absolute path to the installation root (NOT the bin directory).
	// e.g., "C:\Program Files\Java\jdk-17.0.2"
	Path string `json:"path"`
	// Vendor tries to identify the distribution (e.g., "Oracle", "OpenJDK", "Temurin").
	Vendor string `json:"vendor"`
}

func (j JavaInstallation) String() string {
	return fmt.Sprintf("[%s] %s (%s) @ %s", j.Vendor, j.Version, j.MajorVersion, j.Path)
}
