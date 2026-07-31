package update

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// extractTarGz extracts the binary from a .tar.gz archive and replaces currentExe.
func extractTarGz(archivePath, tmpDir, currentExe string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("gzip: %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		name := filepath.Base(header.Name)
		if name != "aimap" {
			continue
		}

		newBinary := filepath.Join(tmpDir, "aimap")
		out, err := os.OpenFile(newBinary, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, tr); err != nil {
			out.Close()
			return err
		}
		out.Close()

		return replaceBinary(newBinary, currentExe)
	}

	return fmt.Errorf("binary not found in archive")
}

// extractZip extracts the binary from a .zip archive and replaces currentExe.
func extractZip(archivePath, tmpDir, currentExe string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		name := filepath.Base(f.Name)
		if name != "aimap.exe" {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		newBinary := filepath.Join(tmpDir, "aimap.exe")
		out, err := os.OpenFile(newBinary, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, rc); err != nil {
			out.Close()
			return err
		}
		out.Close()
		rc.Close()

		return replaceBinary(newBinary, currentExe)
	}

	return fmt.Errorf("binary not found in archive")
}

// replaceBinary replaces currentExe with newBinary.
func replaceBinary(newBinary, currentExe string) error {
	if err := os.Chmod(newBinary, 0755); err != nil {
		return fmt.Errorf("chmod: %w", err)
	}

	if strings.HasSuffix(strings.ToLower(currentExe), ".exe") {
		oldPath := currentExe + ".old"
		os.Remove(oldPath)
		if err := os.Rename(currentExe, oldPath); err != nil {
			return fmt.Errorf("rename current: %w", err)
		}
		if err := os.Rename(newBinary, currentExe); err != nil {
			os.Rename(oldPath, currentExe)
			return fmt.Errorf("rename new: %w", err)
		}
		os.Remove(oldPath)
		return nil
	}

	src, err := os.Open(newBinary)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.OpenFile(currentExe, os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("open current: %w", err)
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

// downloadFile downloads a URL to the given file path.
func downloadFile(url, dest string) error {
	resp, err := httpGet(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
