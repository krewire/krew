package commands

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	kiwver "github.com/krewire/krew/core/version"
)

func downloadAndInstallAsset(ctx context.Context, url, assetName, destPath string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "kiw-self-update/"+kiwver.VersionString())

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned HTTP %d", resp.StatusCode)
	}

	tempDir, err := os.MkdirTemp("", "kiw-self-update-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	archivePath := filepath.Join(tempDir, assetName)
	f, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		return err
	}
	f.Close()

	extractedBinary := filepath.Join(tempDir, "kiw-extracted")
	if strings.HasSuffix(assetName, ".tar.gz") || strings.HasSuffix(assetName, ".tgz") {
		if err := extractTarGzBinary(archivePath, extractedBinary); err != nil {
			return err
		}
	} else if strings.HasSuffix(assetName, ".zip") {
		if err := extractZipBinary(archivePath, extractedBinary); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unsupported archive format: %s", assetName)
	}

	return replaceBinary(extractedBinary, destPath)
}

func extractTarGzBinary(archivePath, destBinary string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		base := filepath.Base(hdr.Name)
		if base == "krew" || base == "krew.exe" || base == "kiw" || base == "kiw.exe" {
			out, err := os.OpenFile(destBinary, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return err
			}
			out.Close()
			return nil
		}
	}
	return fmt.Errorf("binary 'krew' not found in archive")
}

func extractZipBinary(archivePath, destBinary string) error {
	zr, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer zr.Close()

	for _, file := range zr.File {
		base := filepath.Base(file.Name)
		if base == "krew" || base == "krew.exe" || base == "kiw" || base == "kiw.exe" {
			rc, err := file.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			out, err := os.OpenFile(destBinary, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, rc); err != nil {
				out.Close()
				return err
			}
			out.Close()
			return nil
		}
	}
	return fmt.Errorf("binary 'krew' not found in zip archive")
}

func replaceBinary(newBinary, targetPath string) error {
	if err := os.Chmod(newBinary, 0755); err != nil {
		return err
	}

	// Try atomic rename in same directory
	targetDir := filepath.Dir(targetPath)
	tempTarget := filepath.Join(targetDir, fmt.Sprintf(".kiw-new-%d", time.Now().UnixNano()))

	in, err := os.Open(newBinary)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(tempTarget, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("cannot write to target directory %s (may need sudo): %w", targetDir, err)
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		os.Remove(tempTarget)
		return err
	}
	out.Close()

	// On Windows, target may be locked by running process; rename existing to .old
	if runtime.GOOS == "windows" {
		oldPath := targetPath + ".old"
		_ = os.Remove(oldPath)
		_ = os.Rename(targetPath, oldPath)
	}

	if err := os.Rename(tempTarget, targetPath); err != nil {
		os.Remove(tempTarget)
		return fmt.Errorf("replacing binary at %s: %w", targetPath, err)
	}
	return nil
}
