package archive

import (
	"compress/gzip"
	"io"
	"os"
)

type TarGz struct {
	reader io.Reader
}

// NewTarGzFile creates a Gziped tarball for the given filename.
func NewTarGzFile(filename string) (Extractor, error) {
	// #nosec G304 -- filename comes from the caller and is opened as a local archive input.
	data, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	return &TarGz{reader: data}, nil
}

// NewTarGz creates a Gziped tarball for the give Reader.
func NewTarGz(r io.Reader) Extractor {
	return &TarGz{reader: r}
}

// Extract extracts a Gziped tarball into dst.
func (t *TarGz) Extract(dst string) error {
	reader, err := gzip.NewReader(t.reader)
	if err != nil {
		return err
	}

	return NewTar(reader).Extract(dst)
}
