package directory_checksum

import (
	"fmt"
	"github.com/spf13/afero"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const emptySha1 = "da39a3ee5e6b4b0d3255bfef95601890afd80709"

type TestingFilesystemObject interface {
	Create(filesystemImpl afero.Fs)
}

func setUpTestingFilesystem(filesystemObjects []TestingFilesystemObject, filesystemImpl afero.Fs) {
	for _, fso := range filesystemObjects {
		fso.Create(filesystemImpl)
	}
}

type TestingFile struct {
	absolutePath string
	content      string
}

type TestingDir struct {
	absolutePath string
}

func (file TestingFile) Create(filesystemImpl afero.Fs) {
	f, _ := filesystemImpl.Create(file.absolutePath)
	f.WriteString(file.content)
	f.Close()
}

func (dir TestingDir) Create(filesystemImpl afero.Fs) {
	filesystemImpl.Mkdir(dir.absolutePath, os.ModePerm)
}

func TestDeterministicResult(t *testing.T) {
	// Tests whether two consecutive runs of ScanDirectory() produce the same output
	tempDir := t.TempDir() // TODO get rid of TempDir?
	testingFilesystem := []TestingFilesystemObject{
		TestingDir{absolutePath: tempDir + "/d"},
		TestingDir{absolutePath: tempDir + "/d/sub"},
		TestingFile{absolutePath: tempDir + "/f", content: "foo"},
	}
	filesystemImpl := afero.NewMemMapFs()
	setUpTestingFilesystem(testingFilesystem, filesystemImpl)

	d1, _ := ScanDirectory(tempDir, filesystemImpl)
	d1.ComputeDirectoryChecksums()
	output1 := d1.PrintChecksums(3)

	d2, _ := ScanDirectory(tempDir, filesystemImpl)
	d2.ComputeDirectoryChecksums()
	output2 := d2.PrintChecksums(3)

	if output1 != output2 {
		t.Fatalf("Outputs differ:\noutput1:\n%s\n\noutput2:\n%s", output1, output2)
	}

	expectedNewlineCount := len(testingFilesystem) + 2 // one for the root dir, one for the trailing newline

	if l := len(strings.Split(output1, "\n")); l != expectedNewlineCount {
		t.Fatalf("Got %d newlines, expected %d", l, expectedNewlineCount)
	}
}

func TestEmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()
	d, _ := ScanDirectory(tempDir, afero.NewOsFs())
	d.ComputeDirectoryChecksums()
	got := d.PrintChecksums(3)
	want := fmt.Sprintf("%s D .\n", emptySha1)
	if got != want {
		t.Fatalf("Got %s, want %s", got, want)
	}
}

func TestSingleFile(t *testing.T) {
	// Tests whether ScanDirectory() produces expected output for a single file of fixed content
	tempDir := t.TempDir()
	testingFilesystem := []TestingFilesystemObject{
		TestingFile{absolutePath: tempDir + "/f", content: "foo"},
	}
	filesystemImpl := afero.NewMemMapFs()
	setUpTestingFilesystem(testingFilesystem, filesystemImpl)

	d, _ := ScanDirectory(tempDir, filesystemImpl)
	d.ComputeDirectoryChecksums()
	got := d.PrintChecksums(1)

	want := "beb8daa61290acf19e174a689715f32c51b644b6 D .\n" +
		"0beec7b5ea3f0fdbc95d0dd47f3c5bc275da8a33 F f\n"

	if got != want {
		t.Fatalf("Got\n%s\n\nwant\n%s", got, want)
	}
}

func TestSingleDir(t *testing.T) {
	// Tests whether ScanDirectory() produces expected output for a dir that contains only one empty directory
	tempDir := t.TempDir()
	testingFilesystem := []TestingFilesystemObject{
		TestingDir{absolutePath: tempDir + "/dir"},
	}
	filesystemImpl := afero.NewMemMapFs()
	setUpTestingFilesystem(testingFilesystem, filesystemImpl)

	d, _ := ScanDirectory(tempDir, filesystemImpl)
	d.ComputeDirectoryChecksums()
	got := d.PrintChecksums(1)

	want := "365f7001add79c757b245c386b444aca93a73d40 D .\n" +
		"da39a3ee5e6b4b0d3255bfef95601890afd80709 D dir\n"

	if got != want {
		t.Fatalf("Got\n%s\n\nwant\n%s", got, want)
	}
}

func TestLimitingMaxDepth(t *testing.T) {
	// Tests that the output when limiting maximum depth has the expected number of entries
	testingFilesystem := []TestingFilesystemObject{
		TestingDir{absolutePath: filepath.FromSlash("/d")},                         // level 1
		TestingDir{absolutePath: filepath.FromSlash("/d/sub")},                     // level 2
		TestingFile{absolutePath: filepath.FromSlash("/d/sub/f"), content: "foo"},  // level 3
		TestingFile{absolutePath: filepath.FromSlash("/d/sub/f2"), content: "foo"}, // level 3
	}
	filesystemImpl := afero.NewMemMapFs()
	setUpTestingFilesystem(testingFilesystem, filesystemImpl)
	root := string(os.PathSeparator)

	d, _ := ScanDirectory(root, filesystemImpl)
	d.ComputeDirectoryChecksums()

	outputDepth0 := d.PrintChecksums(0)
	gotLinesDepth0 := len(strings.Split(outputDepth0, "\n"))
	if gotLinesDepth0 != 2 {
		t.Fatalf("For depth 0, got %d lines, but expected 2", gotLinesDepth0)
	}

	outputDepth1 := d.PrintChecksums(1)
	gotLinesDepth1 := len(strings.Split(outputDepth1, "\n"))
	if gotLinesDepth1 != 3 {
		t.Fatalf("For depth 1, got %d lines, but expected 3", gotLinesDepth1)
	}

	outputDepth2 := d.PrintChecksums(2)
	gotLinesDepth2 := len(strings.Split(outputDepth2, "\n"))
	if gotLinesDepth2 != 4 {
		t.Fatalf("For depth 1, got %d lines, but expected 4", gotLinesDepth2)
	}
}

func TestNonExistingDirectory(t *testing.T) {
	filesystemImpl := afero.NewMemMapFs()
	_, err := ScanDirectory("/does/not/exist", filesystemImpl)

	if err == nil {
		t.Fatal("Expected error but did not get any")
	}
	if !strings.Contains(err.Error(), "file does not exist") {
		t.Fatalf("Unexpected error was returned: %v", err)
	}
}

func TestFilePath(t *testing.T) {
	filesystemImpl := afero.NewOsFs()
	root := filepath.Join(t.TempDir(), "file")
	f, _ := filesystemImpl.Create(root)
	f.Close()
	_, err := ScanDirectory(root, filesystemImpl)
	if err == nil {
		t.Fatal("Expected error but did not get any")
	}
	if !strings.Contains(err.Error(), "root path must point to a directory") {
		t.Fatalf("Unexpected error was returned: %v", err)
	}
}

func TestDotRoot(t *testing.T) {
	filesystemImpl := afero.NewOsFs()
	d, _ := ScanDirectory(".", filesystemImpl)
	d.ComputeDirectoryChecksums()
	got := d.PrintChecksums(1)
	wantLines := len(strings.Split(got, "\n"))
	if wantLines < 5 {
		t.Fatalf("Expected at least 5 lines of output, but got this:\n%s", got)
	}
}

func TestScanWithUnreadableFile(t *testing.T) {
	testingFilesystem := []TestingFilesystemObject{
		TestingDir{absolutePath: filepath.FromSlash("/d")},    // level 1
		TestingFile{absolutePath: filepath.FromSlash("/d/f")}, // level 2
	}
	filesystemImpl := afero.NewMemMapFs()
	setUpTestingFilesystem(testingFilesystem, filesystemImpl)
	root := string(os.PathSeparator)

	wrapper := fsWrapper{filesystemImpl}
	filesystemImpl = &wrapper

	_, err := ScanDirectory(root, filesystemImpl)

	if err == nil {
		t.Fatal("Expected error but did not get any")
	}
}
