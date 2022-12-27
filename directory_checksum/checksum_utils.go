package directory_checksum

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/go-errors/errors"
	"github.com/spf13/afero"
	"io"
)

// computeFileChecksum computes the SHA-1 digest of the file located at absoluteFilePath and returns it as string that
// represents the digest with hexadecimal notation. If isSymbolicLink is true, the hash is instead computed on the
// link's target, which is basically the "content" of a symbolic link file.
func computeFileChecksum(absoluteFilePath string, isSymbolicLink bool, filesystemImpl afero.Fs) (string, error) {
	if isSymbolicLink {
		linkReader, ok := filesystemImpl.(afero.LinkReader)
		if !ok {
			return "", errors.Errorf("unable to compute checksum for symbolic link file %s: file system is "+
				"unable to read links", absoluteFilePath)
		}

		linkTarget, err := linkReader.ReadlinkIfPossible(absoluteFilePath)
		if err != nil {
			return "", errors.Wrap(err, 0)
		}

		hasher := sha1.New()
		_, err = io.WriteString(hasher, linkTarget)
		if err != nil {
			return "", errors.Wrap(err, 0)
		}
		return hex.EncodeToString(hasher.Sum(nil)), nil
	} else {
		f, err := filesystemImpl.Open(absoluteFilePath)
		if err != nil {
			return "", errors.Wrap(err, 0)
		}
		defer func(f afero.File) {
			err := f.Close()
			if err != nil {
				fmt.Printf("Unable to close file %s in computeFileChecksum(): %v\n", absoluteFilePath, err)
			}
		}(f)

		h := sha1.New()
		if _, err := io.Copy(h, f); err != nil {
			return "", errors.Wrap(err, 0)
		}

		return hex.EncodeToString(h.Sum(nil)), nil
	}
}
