package archive

import (
	"archive/tar"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Tarball struct {
	reader io.Reader
}

// NewTarFile creates a tarball from a given filename.
func NewTarFile(filename string) (Extractor, error) {
	// #nosec G304
	data, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	return &Tarball{reader: data}, nil
}

// NewTar creates a tarball from a given a Reader.
func NewTar(r io.Reader) Extractor {
	return &Tarball{reader: r}
}

// Extract extracts content of a tarball into dst.
func (t *Tarball) Extract(dst string) error {
	if err := os.MkdirAll(dst, 0750); err != nil {
		return err
	}

	tarReader := tar.NewReader(t.reader)

	for {
		header, err := tarReader.Next()
		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		case header == nil:
			continue
		}

		path, err := secureJoin(dst, header.Name)
		if err != nil {
			return err
		}
		info := header.FileInfo()

		switch header.Typeflag {
		case tar.TypeDir:
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return err
			}
		case tar.TypeReg:
			if err = os.MkdirAll(filepath.Dir(path), 0750); err != nil {
				return err
			}
			// #nosec G304
			file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
			switch {
			case os.IsExist(err):
				continue
			case err != nil:
				return err
			}

			if header.Size < 0 {
				if closeErr := file.Close(); closeErr != nil {
					return closeErr
				}
				return errors.New("invalid tar entry size")
			}
			// #nosec G110
			if _, err = io.CopyN(file, tarReader, header.Size); err != nil && !errors.Is(err, io.EOF) {
				if closeErr := file.Close(); closeErr != nil {
					return closeErr
				}
				return err
			}
			if err := file.Close(); err != nil {
				return err
			}
		case tar.TypeLink:
			link, err := secureJoin(dst, header.Name)
			if err != nil {
				return err
			}
			linkTarget, err := secureJoin(dst, header.Linkname)
			if err != nil {
				return err
			}
			if err := os.Link(linkTarget, link); err != nil && !os.IsExist(err) {
				return err
			}
		case tar.TypeSymlink:
			linkPath, err := secureJoin(dst, header.Name)
			if err != nil {
				return err
			}
			if filepath.IsAbs(header.Linkname) {
				return errors.New("absolute symlink target is not allowed")
			}
			if err := os.Symlink(header.Linkname, linkPath); err != nil {
				if !os.IsExist(err) {
					return err
				}
			}
		}
	}
}

func secureJoin(base, name string) (string, error) {
	cleanBase := filepath.Clean(base)
	cleanTarget := filepath.Join(cleanBase, filepath.Clean(name))
	rel, err := filepath.Rel(cleanBase, cleanTarget)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", errors.New("archive entry escapes target directory")
	}
	return cleanTarget, nil
}
