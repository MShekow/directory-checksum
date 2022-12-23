package directory_checksum

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/spf13/afero"
	"io"
	"runtime/debug"
)

// computeFileChecksum computes the SHA-1 digest of the file located at absoluteFilePath and returns it as string that
// represents the digest with hexadecimal notation. If isSymbolicLink is true, the hash is instead computed on the
// link's target, which is basically the "content" of a symbolic link file.
func computeFileChecksum(absoluteFilePath string, isSymbolicLink bool, filesystemImpl afero.Fs) (string, error) {
	if isSymbolicLink {
		linkReader, ok := filesystemImpl.(afero.LinkReader)
		if !ok {
			return "", fmt.Errorf("unable to compute checksum for symbolic link file %s: file system is "+
				"unable to read links", absoluteFilePath)
		}

		linkTarget, err := linkReader.ReadlinkIfPossible(absoluteFilePath)
		if err != nil {
			debug.PrintStack()
			return "", err
		}

		hasher := sha1.New()
		_, err = io.WriteString(hasher, linkTarget)
		if err != nil {
			debug.PrintStack()
			return "", err
		}
		return hex.EncodeToString(hasher.Sum(nil)), nil
	} else {
		f, err := filesystemImpl.Open(absoluteFilePath)
		if err != nil {
			return "", err
		}
		defer f.Close()

		h := sha1.New()
		if _, err := io.Copy(h, f); err != nil {
			return "", err
		}

		return hex.EncodeToString(h.Sum(nil)), nil
	}
}
