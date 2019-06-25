# gopacker
A UPX-like packer to shrink executables.

## Quick Start
```
go get github.com/nirhaas/gopacker
gopacker <executable_to_pack>
```

## How does it work

### Packing
* Copy `gopacker` executalbe itself to output file.
* Compress and stream (append) to output file.
* Append compressed size.
* Append magic string.

Output file is now a functional executable.

### Unpacking
When running the packed executable:
* Checks the last few bytes to see if magic string is there.
* Reading compressed size.
* Reading compressed data.
* Uncompressing to memory.
* Overriding the packed executable.
* syscall exec to run the unpacked executable.

Possible TODO:
* Better compression.
* Encryption.