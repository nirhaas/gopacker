package lib

import (
	"io"
)

// Compression is an interface to be implemented to standatrize compression
// mechanisms used by this tool.
type Compression interface {
	CompressWriter(io.Writer) (io.WriteCloser, error)
	DecompressReader(io.Reader) (io.ReadCloser, error)
}

// CompressStream compress by streaming from in to out.
func CompressStream(compression Compression, out io.Writer, in io.Reader) (n int64, err error) {
	writer, err := compression.CompressWriter(out)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err == nil {
			err = writer.Close()
		}
	}()
	return io.Copy(writer, in)
}

// DecompressStream compress by streaming from in to out.
func DecompressStream(compression Compression, out io.Writer, in io.Reader) (n int64, err error) {
	reader, err := compression.DecompressReader(in)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err == nil {
			err = reader.Close()
		}
	}()
	return io.Copy(out, reader)
}
