package lib

import (
	"compress/gzip"
	"io"
)

// GZIPCompression .
type GZIPCompression struct{}

// CompressWriter .
func (z GZIPCompression) CompressWriter(out io.Writer) (io.WriteCloser, error) {
	return gzip.NewWriterLevel(out, gzip.BestSpeed)
}

// DecompressReader .
func (z GZIPCompression) DecompressReader(in io.Reader) (io.ReadCloser, error) {
	return gzip.NewReader(in)
}
