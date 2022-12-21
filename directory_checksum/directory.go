package directory_checksum

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
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
func (d *Directory) ComputeDirectoryChecksums() string {
	hasher := sha1.New()

	for _, dirName := range sortedKeys(d.dirs) {
		childDir := d.dirs[dirName]
		childDirChecksum := childDir.ComputeDirectoryChecksums()
		io.WriteString(hasher, fmt.Sprintf("'%s' %s\n", dirName, childDirChecksum))
	}
	for _, fileName := range sortedKeys(d.files) {
		childFileChecksum := d.files[fileName]
		io.WriteString(hasher, fmt.Sprintf("'%s' %s\n", fileName, childFileChecksum))
	}

	d.checksum = hex.EncodeToString(hasher.Sum(nil))

	return d.checksum
}

// PrintChecksums prints a listing of the files and directories, including their checksums, using pre-order tree
// traversal, stopping the traversal at the specified depth level. It assumes that ComputeDirectoryChecksums() has
// already been called on the root Directory object.
func (d *Directory) PrintChecksums(relativePath string, depth int) string {
	stringBuilder := strings.Builder{}
	stringBuilder.WriteString(fmt.Sprintf("%s D %s\n", d.checksum, relativePath))
	if depth <= 0 {
		return stringBuilder.String()
	}

	for _, dirName := range sortedKeys(d.dirs) {
		stringBuilder.WriteString(d.dirs[dirName].PrintChecksums(filepath.Join(relativePath, dirName), depth-1))
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
func (d *Directory) Add(relativeRemainingPath string, relativePath string, absoluteRootPath string, isDir bool) {
	if strings.Contains(relativeRemainingPath, "/") {
		components := strings.SplitN(relativeRemainingPath, "/", 2)
		subDir := d.dirs[components[0]]
		subDir.Add(components[1], relativePath, absoluteRootPath, isDir)
	} else {
		if isDir {
			d.dirs[relativeRemainingPath] = newDirectory()
		} else {
			absoluteFilePath := filepath.Join(absoluteRootPath, relativePath)
			d.files[relativeRemainingPath] = computeChecksum(absoluteFilePath)
		}
	}
}
