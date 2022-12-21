package directory_checksum

import (
	"crypto/sha1"
	"encoding/hex"
	"github.com/spf13/afero"
	"io"
)

// computeChecksum computes the SHA-1 digest of the file located at absoluteFilePath and returns it as string that
// represents the digest with hexadecimal notation.
func computeChecksum(absoluteFilePath string, filesystemImpl afero.Fs) (string, error) {
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
