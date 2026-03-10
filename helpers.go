package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ListDirs recursively walks through 'dir' and returns a slice of all directories
func ListDirs(dir string) ([]string, error) {
	var dirs []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("[DataHub.main.ListDirs] os.ReadDir error: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}

	return dirs, err
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("[DataHub.main.copyFile] os.Open error: %w", err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("[DataHub.main.copyFile] os.Create error: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func extractZIP(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("[DataHub.main.extractZIP] zip.OpenReader error: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, 0755)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
			return fmt.Errorf("[DataHub.main.extractZIP] os.MkdirAll error: %w", err)
		}
		outFile, err := os.Create(fpath)
		if err != nil {
			return fmt.Errorf("[DataHub.main.extractZIP] os.Create error: %w", err)
		}
		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}
		if _, err = io.Copy(outFile, rc); err != nil {
			outFile.Close()
			rc.Close()
			return err
		}
		outFile.Close()
		rc.Close()
	}
	return nil
}

func extractTAR(src, dest string) error {
	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("[DataHub.main.extractTAR] os.Open error: %w", err)
	}
	defer f.Close()
	return untar(f, dest)
}

func extractTARGZ(src, dest string) error {
	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("[DataHub.main.extractTARGZ] os.Open error: %w", err)
	}
	defer f.Close()
	gzr, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("[DataHub.main.extractTARGZ] gzip.NewReader error: %w", err)
	}
	defer gzr.Close()
	return untar(gzr, dest)
}

func untar(r io.Reader, dest string) error {
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("[DataHub.main.untar] tr.Next error: %w", err)
		}

		fpath := filepath.Join(dest, hdr.Name)
		switch hdr.Typeflag {
		case tar.TypeDir:
			os.MkdirAll(fpath, 0755)
		case tar.TypeReg:
			os.MkdirAll(filepath.Dir(fpath), 0755)
			outFile, err := os.Create(fpath)
			if err != nil {
				return fmt.Errorf("[DataHub.main.untar] os.Create error: %w", err)
			}
			if _, err = io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}
	return nil
}

func errMissingParam(param string) error {
	return &paramError{msg: "missing parameter: " + param}
}

func errNotFoundDir(dir string) error {
	return &paramError{msg: "directory not found: " + dir}
}

type paramError struct {
	msg string
}

func (e *paramError) Error() string { return e.msg }
