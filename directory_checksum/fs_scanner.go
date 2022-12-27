package directory_checksum

import (
	"github.com/go-errors/errors"
	"github.com/spf13/afero"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// bendRelativePath converts the provided relativePath (which MIGHT actually be absolute) to an actually relative path,
// that is relative to absoluteRootPath.
func bendRelativePath(relativePath, absoluteRootPath string) string {
	if strings.HasPrefix(relativePath, absoluteRootPath) {
		/*
			For absoluteRootPath = "/foo" and relativePath "/foo/bar", we need an extraLengthCutoff of 1, to turn
			relativePath into "bar".

			But for absoluteRootPath = "/" and relativePath "/foo/bar", extraLengthCutoff must be 0, to avoid the bad
			result "oo/bar".
		*/
		isAbsRoot := absoluteRootPath == "/" || absoluteRootPath == "\\"
		extraLengthCutoff := 1
		if isAbsRoot {
			extraLengthCutoff = 0
		}
		relativePath = relativePath[len(absoluteRootPath)+extraLengthCutoff:]
	}
	return relativePath
}

// ScanDirectory returns the pointer to a (hierarchically-nested) Directory that is constructed from recursively walking
// the directory located at absoluteRootPath.
func ScanDirectory(absoluteRootPath string, filesystemImpl afero.Fs) (*Directory, error) {
	// Handle a special case that happens only during unit testing (where root is '\' when executed on Windows)
	if absoluteRootPath != "\\" {
		absRootPath, err := filepath.Abs(absoluteRootPath)
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
		absoluteRootPath = absRootPath
	}

	directory := newDirectory()
	err := afero.Walk(filesystemImpl, absoluteRootPath, func(relativePath string, info fs.FileInfo, err error) error {
		if err != nil {
			return errors.Wrap(err, 0)
		}
		// Walk() is happy to walk a FILE (instead of a dir) -> we have to manually check that a dir path was provided
		if absoluteRootPath == relativePath && !info.IsDir() {
			return errors.New("provided root path must point to a directory")
		}

		if relativePath != absoluteRootPath {
			relativePath = bendRelativePath(relativePath, absoluteRootPath)

			fileType := TypeFile
			if info.IsDir() {
				fileType = TypeDir
			} else if info.Mode()&os.ModeSymlink == os.ModeSymlink {
				fileType = TypeSymlink
			}

			err := directory.Add(relativePath, relativePath, absoluteRootPath, fileType, filesystemImpl)
			if err != nil {
				return errors.Wrap(err, 0)
			}
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}

	return directory, nil
}
