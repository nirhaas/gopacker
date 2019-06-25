package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/nirhaas/gopacker/lib"

	"github.com/pkg/errors"
)

const (
	usageString       = "USAGE: packer <executable_path>"
	packedExt         = ".packed"
	footerMagicString = "LALALALA"
)

var (
	// Compression method.
	compression = lib.ZSTDCompression{}
	// Placeholders.
	footerMagic  = []byte(footerMagicString)
	selfPath     string
	inputPath    string
	packedPath   string
	unpackedPath string
)

// Writer that counts how many bytes were written. User here to get compressed
// size when streaming with io.Copy.
type counterWriter struct {
	io.Writer
	n int
}

func (w *counterWriter) Write(buf []byte) (n int, err error) {
	n, err = w.Writer.Write(buf)
	w.n += n
	return n, err
}

func appendToFile(dst string, in io.Reader, compress bool) (n int64, err error) {
	// Check that destination file exists.
	_, err = os.Stat(dst)
	if err != nil {
		return 0, errors.Wrap(err, "destination file does not exist")
	}

	// Open file on APPEND mode.
	destination, err := os.OpenFile(dst, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return 0, errors.Wrap(err, "failed opening destination file")
	}
	defer func() {
		if err == nil {
			err = destination.Close()
		}
	}()

	// Compress and stream.
	if compress {
		cw := &counterWriter{Writer: destination}
		_, err := lib.CompressStream(compression, cw, in)
		return int64(cw.n), errors.Wrap(err, "failed appending compressed to destination")
	}

	// Just stream.
	return io.Copy(destination, in)

}

func copyFile(src, dst string) (err error) {
	// Verify that source file exists.
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return errors.Wrap(err, "src file does not exist")
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	// Open source file.
	source, err := os.Open(src)
	if err != nil {
		return errors.Wrap(err, "failed opening source file")
	}
	defer func() { _ = source.Close() }() // Best effort.

	// Open dest file.
	destination, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		return errors.Wrap(err, "failed opening destination file")
	}
	defer func() {
		if err == nil {
			err = destination.Close()
		}
	}()

	// Copy.
	_, err = io.Copy(destination, source)
	return err
}

// CLI tool.
func mainCLI() {
	// Parse arg.
	if len(os.Args) == 1 {
		fmt.Println(usageString)
		return
	}

	inputPath = os.Args[1]
	packedPath = os.Args[1] + packedExt

	// See that the file exists.
	if _, err := os.Stat(os.Args[1]); err != nil {
		log.Fatal(errors.Wrap(err, "target exec does not exist"))
	}

	// Copy self (stub) to final path.
	if err := copyFile(selfPath, packedPath); err != nil {
		log.Fatal(errors.Wrap(err, "failed copying stub"))
	}

	// Open packed.
	packedPathHandle, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed opening packed path"))
	}

	// Compress and append to final path.
	bytesWritten, err := appendToFile(packedPath, packedPathHandle, true)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed appending stub"))
	}

	// Append footer.
	bs := make([]byte, 8)
	binary.LittleEndian.PutUint64(bs, uint64(bytesWritten))
	if _, err := appendToFile(packedPath, bytes.NewReader(bs), false); err != nil {
		log.Fatal(errors.Wrap(err, "failed appending len"))
	}

	if _, err := appendToFile(packedPath, bytes.NewReader(footerMagic), false); err != nil {
		log.Fatal(errors.Wrap(err, "failed appending magic"))
	}

}

// This is a stub. Unpack and run.
func mainStub(selfFileHandle *os.File, fileSize int64) {
	unpackedPath = selfPath

	// Read DWORD before the magic, which is the data length to extract.
	dstLen := make([]byte, 8)
	if _, err := selfFileHandle.Seek(fileSize-int64(len(footerMagic))-int64(len(dstLen)), 0); err != nil {
		log.Fatal(errors.Wrap(err, "failed seeking length"))
	}
	if _, err := selfFileHandle.Read(dstLen); err != nil {
		log.Fatal(errors.Wrap(err, "failed reading length"))
	}
	targetLen := int64(binary.LittleEndian.Uint64(dstLen))

	// Decompress all to memory. We don't stream here so we can overwrite the current
	// executable. Streaming where src and dst are equal just breaks everything.
	var buf bytes.Buffer
	compressedOff := fileSize - int64(len(footerMagic)) - int64(len(dstLen)) - targetLen
	compressedReader := io.NewSectionReader(selfFileHandle, compressedOff, targetLen)
	if _, err := lib.DecompressStream(compression, &buf, compressedReader); err != nil {
		log.Fatal(errors.Wrap(err, "failed writing unpacked file"))
	}

	// Write to file (overwrite).
	if err := ioutil.WriteFile(unpackedPath, buf.Bytes(), 0755); err != nil {
		log.Fatal(errors.Wrap(err, "failed writing unpacked file"))
	}

	// Exec.
	if err := syscall.Exec(unpackedPath, []string{unpackedPath}, os.Environ()); err != nil {
		log.Fatal(errors.Wrap(err, "failed exec-ing to unpacked executable"))
	}
}

func main() {
	// No := so selfPath won't be overridden.
	var err error
	selfPath, err = exec.LookPath(os.Args[0])
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed expanding self path"))
	}

	// Get self (stub) executable path.
	selfFileStat, err := os.Stat(selfPath)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed stat self file"))
	}

	// Open self.
	selfFileHandle, err := os.Open(selfPath)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed opening self file"))
	}
	defer func() { _ = selfFileHandle.Close() }() // Best effort.

	// Read the end of the file to get the magic.
	dstMagic := make([]byte, len(footerMagic))
	if _, err := selfFileHandle.Seek(selfFileStat.Size()-int64(len(footerMagic)), 0); err != nil {
		log.Fatal(errors.Wrap(err, "failed seeking magic"))
	}
	if _, err := selfFileHandle.Read(dstMagic); err != nil {
		log.Fatal(errors.Wrap(err, "failed reading magic"))
	}

	// Compare found magic with expected magic.
	if !bytes.Equal(footerMagic, dstMagic) {
		// Magic not found - CLI tool.
		mainCLI()
		return
	}

	// Magic found - this is a stub.
	mainStub(selfFileHandle, selfFileStat.Size())
	return
}
