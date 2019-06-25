package lib

import (
	"io"

	"github.com/klauspost/compress/zstd"
)

// ZSTDReaderWithCloser is a wrapper because original zstd.Decoder is not
// implementing io.ReadCloser correctly (does not return error).
type ZSTDReaderWithCloser struct {
	r *zstd.Decoder
}

// Read wraps zstd.Decoder Read.
func (zr ZSTDReaderWithCloser) Read(b []byte) (int, error) {
	return zr.r.Read(b)
}

// Close just wraps zstd.Decoder Close, but return nil as error.
func (zr ZSTDReaderWithCloser) Close() error {
	zr.r.Close()
	return nil
}

// NewZSTDReaderWithCloser .
func NewZSTDReaderWithCloser(reader *zstd.Decoder) ZSTDReaderWithCloser {
	return ZSTDReaderWithCloser{
		r: reader,
	}
}

// ZSTDCompression .
type ZSTDCompression struct{}

// CompressWriter .
func (z ZSTDCompression) CompressWriter(out io.Writer) (io.WriteCloser, error) {
	return zstd.NewWriter(out, zstd.WithEncoderLevel(zstd.SpeedBetterCompression))
}

// DecompressReader .
func (z ZSTDCompression) DecompressReader(in io.Reader) (io.ReadCloser, error) {
	reader, err := zstd.NewReader(in)
	return NewZSTDReaderWithCloser(reader), err
}
