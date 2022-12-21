package directory_checksum

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/spf13/afero"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// A Directory represents a physical directory on the file system. files and dirs contain only the immediate child
// objects. The files field maps from the file's name to its SHA-1 checksum. The dirs field maps from the directory's
// name to its corresponding Directory pointer.
type Directory struct {
	files    map[string]string
	dirs     map[string]*Directory
	checksum string
}

// newDirectory constructs an empty Directory object with pre-initialized empty maps.
func newDirectory() *Directory {
	d := Directory{
		files:    map[string]string{},
		dirs:     map[string]*Directory{},
		checksum: "",
	}
	return &d
}

// ComputeDirectoryChecksums recursively computes the "checksum" field of all Directory objects, and returns the
// checksum of the object this method is called on.
// It assumes that the checksum of all files(!) have already been computed.
func (d *Directory) ComputeDirectoryChecksums() (string, error) {
	hasher := sha1.New()

	for _, dirName := range sortedKeys(d.dirs) {
		childDir := d.dirs[dirName]
		childDirChecksum, err := childDir.ComputeDirectoryChecksums()
		if err != nil {
			return "", err
		}
		_, err = io.WriteString(hasher, fmt.Sprintf("'%s' %s\n", dirName, childDirChecksum))
		if err != nil {
			return "", err
		}
	}
	for _, fileName := range sortedKeys(d.files) {
		childFileChecksum := d.files[fileName]
		_, err := io.WriteString(hasher, fmt.Sprintf("'%s' %s\n", fileName, childFileChecksum))
		if err != nil {
			return "", err
		}
	}

	d.checksum = hex.EncodeToString(hasher.Sum(nil))

	return d.checksum, nil
}

// PrintChecksums prints a listing of the files and directories, including their checksums, using pre-order tree
// traversal, stopping the traversal at the specified depth level. It assumes that ComputeDirectoryChecksums() has
// already been called on the root Directory object.

func (d *Directory) PrintChecksums(depth int) string {
	return d.printChecksums(".", depth)
}

// printChecksums is the actual implementation of PrintChecksums.
func (d *Directory) printChecksums(relativePath string, depth int) string {
	stringBuilder := strings.Builder{}
	stringBuilder.WriteString(fmt.Sprintf("%s D %s\n", d.checksum, relativePath))
	if depth <= 0 {
		return stringBuilder.String()
	}

	for _, dirName := range sortedKeys(d.dirs) {
		stringBuilder.WriteString(d.dirs[dirName].printChecksums(filepath.Join(relativePath, dirName), depth-1))
	}

	for _, fileName := range sortedKeys(d.files) {
		fileChecksum := d.files[fileName]
		stringBuilder.WriteString(fmt.Sprintf("%s F %s\n", fileChecksum, filepath.Join(relativePath, fileName)))
	}

	return stringBuilder.String()
}

// Add adds the file or directory located at absoluteRootPath/relativePath to the correct Directory object.
// relativeRemainingPath is a helper argument used to traverse down the Directory object hierarchy, and must initially
// be set to the same value as relativePath. If isDir is false, the SHA-1 checksum is computed
func (d *Directory) Add(relativeRemainingPath string, relativePath string, absoluteRootPath string, isDir bool,
	filesystemImpl afero.Fs) error {
	if strings.Contains(relativeRemainingPath, string(os.PathSeparator)) {
		components := strings.SplitN(relativeRemainingPath, string(os.PathSeparator), 2)
		subDir := d.dirs[components[0]]
		err := subDir.Add(components[1], relativePath, absoluteRootPath, isDir, filesystemImpl)
		if err != nil {
			return err
		}
	} else {
		if isDir {
			d.dirs[relativeRemainingPath] = newDirectory()
		} else {
			absoluteFilePath := filepath.Join(absoluteRootPath, relativePath)
			fileChecksum, err := computeChecksum(absoluteFilePath, filesystemImpl)
			if err != nil {
				return err
			}
			d.files[relativeRemainingPath] = fileChecksum
		}
	}

	return nil
}
