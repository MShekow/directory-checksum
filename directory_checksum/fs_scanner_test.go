package directory_checksum

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

const emptySha1 = "da39a3ee5e6b4b0d3255bfef95601890afd80709"

type TestingFilesystemObject interface {
	Create()
}

func setUpTestingFilesystem(filesystemObjects []TestingFilesystemObject) {
	for _, fso := range filesystemObjects {
		fso.Create()
	}
}

type TestingFile struct {
	absolutePath string
	content      string
}

type TestingDir struct {
	absolutePath string
}

func (file TestingFile) Create() {
	f, _ := os.Create(file.absolutePath)
	f.WriteString(file.content)
	f.Close()
}

func (dir TestingDir) Create() {
	os.Mkdir(dir.absolutePath, os.ModePerm)
}

func TestDeterministicResult(t *testing.T) {
	// Tests whether two consecutive runs of ScanDirectory() produce the same output
	tempDir := t.TempDir()
	testingFilesystem := []TestingFilesystemObject{
		TestingDir{absolutePath: tempDir + "/d"},
		TestingDir{absolutePath: tempDir + "/d/sub"},
		TestingFile{absolutePath: tempDir + "/f", content: "foo"},
	}
	setUpTestingFilesystem(testingFilesystem)

	d1 := ScanDirectory(tempDir)
	d1.ComputeDirectoryChecksums()
	output1 := d1.PrintChecksums(tempDir, 3)

	d2 := ScanDirectory(tempDir)
	d2.ComputeDirectoryChecksums()
	output2 := d2.PrintChecksums(tempDir, 3)

	if output1 != output2 {
		t.Errorf("Outputs differ:\noutput1:\n%s\n\noutput2:\n%s", output1, output2)
	}

	expectedNewlineCount := len(testingFilesystem) + 2 // one for the root dir, one for the trailing newline

	if l := len(strings.Split(output1, "\n")); l != expectedNewlineCount {
		t.Errorf("Got %d newlines, expected %d", l, expectedNewlineCount)
	}
}

func TestEmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()
	d := ScanDirectory(tempDir)
	d.ComputeDirectoryChecksums()
	got := d.PrintChecksums(tempDir, 3)
	want := fmt.Sprintf("%s D %s\n", emptySha1, tempDir)
	if got != want {
		t.Errorf("Got %s, want %s", got, want)
	}
}

func TestSingleFile(t *testing.T) {
	// Tests whether ScanDirectory() produces expected output for a single file of fixed content
	tempDir := t.TempDir()
	testingFilesystem := []TestingFilesystemObject{
		TestingFile{absolutePath: tempDir + "/f", content: "foo"},
	}
	setUpTestingFilesystem(testingFilesystem)

	d := ScanDirectory(tempDir)
	d.ComputeDirectoryChecksums()
	got := d.PrintChecksums(tempDir, 1)

	want := fmt.Sprintf("beb8daa61290acf19e174a689715f32c51b644b6 D %s\n"+
		"0beec7b5ea3f0fdbc95d0dd47f3c5bc275da8a33 F %s%sf\n", tempDir, tempDir, string(os.PathSeparator))

	if got != want {
		t.Errorf("Got\n%s\n\nwant\n%s", got, want)
	}
}

func TestSingleDir(t *testing.T) {
	// Tests whether ScanDirectory() produces expected output for a dir that contains only one empty directory
	tempDir := t.TempDir()
	testingFilesystem := []TestingFilesystemObject{
		TestingDir{absolutePath: tempDir + "/dir"},
	}
	setUpTestingFilesystem(testingFilesystem)

	d := ScanDirectory(tempDir)
	d.ComputeDirectoryChecksums()
	got := d.PrintChecksums(tempDir, 1)

	want := fmt.Sprintf("365f7001add79c757b245c386b444aca93a73d40 D %s\n"+
		"da39a3ee5e6b4b0d3255bfef95601890afd80709 D %s%sdir\n", tempDir, tempDir, string(os.PathSeparator))

	if got != want {
		t.Errorf("Got\n%s\n\nwant\n%s", got, want)
	}
}

func TestLimitingMaxDepth(t *testing.T) {
	// Tests that the output when limiting maximum depth has the expected number of entries
	tempDir := t.TempDir()
	testingFilesystem := []TestingFilesystemObject{
		TestingDir{absolutePath: tempDir + "/d"},                         // level 1
		TestingDir{absolutePath: tempDir + "/d/sub"},                     // level 2
		TestingFile{absolutePath: tempDir + "/d/sub/f", content: "foo"},  // level 3
		TestingFile{absolutePath: tempDir + "/d/sub/f2", content: "foo"}, // level 3
	}
	setUpTestingFilesystem(testingFilesystem)

	d := ScanDirectory(tempDir)
	d.ComputeDirectoryChecksums()

	outputDepth0 := d.PrintChecksums(tempDir, 0)
	gotLinesDepth0 := len(strings.Split(outputDepth0, "\n"))
	if gotLinesDepth0 != 2 {
		t.Errorf("For depth 0, got %d lines, but expected 2", gotLinesDepth0)
	}

	outputDepth1 := d.PrintChecksums(tempDir, 1)
	gotLinesDepth1 := len(strings.Split(outputDepth1, "\n"))
	if gotLinesDepth1 != 3 {
		t.Errorf("For depth 1, got %d lines, but expected 3", gotLinesDepth1)
	}

	outputDepth2 := d.PrintChecksums(tempDir, 2)
	gotLinesDepth2 := len(strings.Split(outputDepth2, "\n"))
	if gotLinesDepth2 != 4 {
		t.Errorf("For depth 1, got %d lines, but expected 4", gotLinesDepth2)
	}
}
