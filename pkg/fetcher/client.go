package fetcher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"
)

const baseURL = "https://api.adoptium.net/v3/assets/feature_releases/%d/ga"

type Release struct {
	Binaries    []Binary    `json:"binaries"`
	VersionData VersionData `json:"version_data"`
}

type VersionData struct {
	Semver string `json:"semver"`
}

type Binary struct {
	Package      Package `json:"package"`
	Architecture string  `json:"architecture"`
	Os           string  `json:"os"`
	ImageType    string  `json:"image_type"`
}

type Package struct {
	Link     string `json:"link"`
	Name     string `json:"name"`
	Checksum string `json:"checksum"`
}

// GetLatestVersion returns the download URL and the semantic version string for the requested major Java version.
func GetLatestVersion(version int) (string, string, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	osParam := getOSParam()
	archParam := getArchParam()

	// Construct URL with query parameters
	url := fmt.Sprintf(baseURL, version)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}

	q := req.URL.Query()
	q.Add("os", osParam)
	q.Add("architecture", archParam)
	q.Add("image_type", "jdk")
	q.Add("jvm_impl", "hotspot")
	q.Add("vendor", "eclipse")
	q.Add("page_size", "1") // We only need the latest one
	q.Add("sort_order", "DESC")

	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var releases []Release
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return "", "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(releases) == 0 {
		return "", "", fmt.Errorf("no releases found for Java %d on %s/%s", version, osParam, archParam)
	}

	// The API returns a list of releases. The first one should be the latest due to sort_order=DESC (default dates) or we just take the first GA.
	// We filtered by feature_releases/{version}/ga so these are GA releases.

	release := releases[0]
	if len(release.Binaries) == 0 {
		return "", "", fmt.Errorf("no binaries found in the latest release")
	}

	// Just take the first binary as we already filtered by OS/Arch
	return release.Binaries[0].Package.Link, release.VersionData.Semver, nil
}

func getOSParam() string {
	switch runtime.GOOS {
	case "darwin":
		return "mac"
	case "windows":
		return "windows"
	case "linux":
		return "linux"
	default:
		return runtime.GOOS // fallback
	}
}

func getArchParam() string {
	switch runtime.GOARCH {
	case "amd64":
		return "x64"
	case "arm64":
		return "aarch64"
	case "386":
		return "x32"
	default:
		return runtime.GOARCH
	}
}
