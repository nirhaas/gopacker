package lib

import (
	"io"

	"github.com/ulikunitz/xz"
	fastxz "github.com/xi2/xz"
)

// XZCompression .
type XZCompression struct{}

func (x XZCompression) CompressWriter(out io.Writer) (io.WriteCloser, error) {
	return xz.NewWriter(out)
}

func (x XZCompression) DecompressReader(in io.Reader) (io.ReadCloser, error) {
	return NewXZReaderWithCloser(fastxz.NewReader(in, 0)), nil
}

// XZReaderWithCloser is a wrapper because original fastxz.Reader is not
// implementing io.ReadCloser correctly (does not return error).
type XZReaderWithCloser struct {
	r *fastxz.Reader
}

// Read wraps fastxz.Reader Read.
func (zr XZReaderWithCloser) Read(b []byte) (int, error) {
	return zr.r.Read(b)
}

// Close just wraps fastxz.Reader Close, but return nil as error.
func (zr XZReaderWithCloser) Close() error {
	// zr.r.Close()
	return nil
}

// NewXZReaderWithCloser .
func NewXZReaderWithCloser(reader *fastxz.Reader, err error) XZReaderWithCloser {
	return XZReaderWithCloser{
		r: reader,
	}
}
