package directory_checksum

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"log"
	"os"
)

// computeChecksum computes the SHA-1 digest of the file located at absoluteFilePath and returns it as string that
// represents the digest with hexadecimal notation.
func computeChecksum(absoluteFilePath string) string {
	f, err := os.Open(absoluteFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
	}

	return hex.EncodeToString(h.Sum(nil))
}
