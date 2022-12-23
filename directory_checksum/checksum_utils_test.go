package directory_checksum

import (
	"github.com/spf13/afero"
	"testing"
)

func TestComputeChecksum(t *testing.T) {
	filesystemImpl := afero.NewMemMapFs()
	tempFilePath := "/tmpfile"
	f, _ := filesystemImpl.Create(tempFilePath)
	f.WriteString("Hello World")
	f.Close()

	got, _ := computeFileChecksum(tempFilePath, false, filesystemImpl)
	want := "0a4d55a8d778e5022fab701977c5d840bbc486d0"

	if got != want {
		t.Fatalf("Got %s, wanted %s", got, want)
	}
}

func TestNonExistingFile(t *testing.T) {
	filesystemImpl := afero.NewMemMapFs()

	_, err := computeFileChecksum("does-not-exist", false, filesystemImpl)

	if err == nil {
		t.Fatal("Expected error but did not get any")
	}
}

func TestUnreadableFile(t *testing.T) {
	filesystemImpl := afero.NewMemMapFs()
	tempFilePath := "/tmpfile"
	f, _ := filesystemImpl.Create(tempFilePath)
	f.WriteString("Hello World")
	f.Close()

	wrapper := fsWrapper{filesystemImpl}
	filesystemImpl = &wrapper

	_, err := computeFileChecksum("/tmpfile", false, filesystemImpl)

	if err == nil {
		t.Fatal("Expected error but did not get any")
	}
}
