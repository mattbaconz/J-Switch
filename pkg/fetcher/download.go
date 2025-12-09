package fetcher

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// ProgressWriter counts the number of bytes written to it. It implements to the io.Writer interface
// and we can pass this into io.TeeReader() which will report progress on each write cycle.
type ProgressWriter struct {
	Total      int64
	Downloaded int64
	OnProgress func(float64)
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n := len(p)
	pw.Downloaded += int64(n)
	if pw.Total > 0 && pw.OnProgress != nil {
		percentage := float64(pw.Downloaded) / float64(pw.Total)
		pw.OnProgress(percentage)
	}
	return n, nil
}

// DownloadAndExtract downloads the file from url and extracts it to destFolder.
// Sends progress (0.0 - 1.0) to progressChan.
func DownloadAndExtract(url string, destFolder string, progressChan chan float64) (string, error) {
	// 1. Download
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	// Create a temp file
	tempFile, err := os.CreateTemp("", "jdk-download-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name()) // Clean up
	defer tempFile.Close()

	// Setup progress tracking
	pw := &ProgressWriter{
		Total: resp.ContentLength,
		OnProgress: func(p float64) {
			// Non-blocking send
			select {
			case progressChan <- p:
			default:
			}
		},
	}

	// Copy from response to temp file, tracking progress
	_, err = io.Copy(tempFile, io.TeeReader(resp.Body, pw))
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	// Ensure 100% is sent
	select {
	case progressChan <- 1.0:
	default:
	}

	// Close file to flush writes before extraction
	tempFile.Close()

	// 2. Extract
	extractedPath, err := extract(tempFile.Name(), destFolder)
	if err != nil {
		return "", fmt.Errorf("extraction failed: %w", err)
	}

	return extractedPath, nil
}

func extract(src string, dest string) (string, error) {
	// Detect file type by signature or just look at content (since we don't have extension easily from temp file unless we preserved it)
	// Actually, API usually gives a .zip or .tar.gz.
	// Let's try to open as zip first, if error, try tar.gz.
	// Or we can peak bytes.

	// For simplicity, let's try opening as zip.
	if isZip(src) {
		return unzip(src, dest)
	}
	return untar(src, dest)
}

func isZip(filename string) bool {
	f, err := os.Open(filename)
	if err != nil {
		return false
	}
	defer f.Close()
	buf := make([]byte, 4)
	if _, err := f.Read(buf); err != nil {
		return false
	}
	return string(buf) == "PK\x03\x04"
}

func unzip(src string, dest string) (string, error) {
	r, err := zip.OpenReader(src)
	if err != nil {
		return "", err
	}
	defer r.Close()

	var rootDir string

	for _, f := range r.File {
		// Store the root directory name
		if rootDir == "" {
			parts := strings.Split(f.Name, "/")
			if len(parts) > 0 {
				rootDir = filepath.Join(dest, parts[0])
			}
		}

		fpath := filepath.Join(dest, f.Name)

		// Check for Zip Slip
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return "", fmt.Errorf("illegal file path: %s", fpath)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return "", err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return "", err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return "", err
		}

		_, err = io.Copy(outFile, rc)

		outFile.Close()
		rc.Close()

		if err != nil {
			return "", err
		}
	}
	return rootDir, nil
}

func untar(src string, dest string) (string, error) {
	f, err := os.Open(src)
	if err != nil {
		return "", err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	var rootDir string

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		// Store root dir
		if rootDir == "" {
			parts := strings.Split(header.Name, "/")
			if len(parts) > 0 {
				rootDir = filepath.Join(dest, parts[0])
			}
		}

		target := filepath.Join(dest, header.Name)

		// Zip Slip check
		if !strings.HasPrefix(target, filepath.Clean(dest)+string(os.PathSeparator)) {
			return "", fmt.Errorf("illegal file path: %s", target)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return "", err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return "", err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return "", err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return "", err
			}
			f.Close()
		}
	}
	return rootDir, nil
}
