package directory_checksum

import (
	"os"
	"testing"
)

func TestComputeChecksum(t *testing.T) {
	tempDir := t.TempDir()
	tempFilePath := tempDir + "/tmpfile"
	f, _ := os.Create(tempFilePath)
	f.WriteString("Hello World")
	f.Close()

	got := computeChecksum(tempFilePath)
	want := "0a4d55a8d778e5022fab701977c5d840bbc486d0"

	if got != want {
		t.Errorf("Got %s, wanted %s", got, want)
	}
}
