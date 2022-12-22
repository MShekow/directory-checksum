package directory_checksum

import (
	"errors"
	"github.com/spf13/afero"
	"io/fs"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
)

// bendRelativePath converts the provided relativePath (which MIGHT actually be absolute) to an actually relative path,
// that is relative to absoluteRootPath.
func bendRelativePath(relativePath, absoluteRootPath string) string {
	if strings.HasPrefix(relativePath, absoluteRootPath) {
		/*
			For absoluteRootPath = "/foo" and relativePath "/foo/bar", we need an extraLengthCutoff of 1, to turn
			relativePath into "bar".

			But for absoluteRootPath = "/" and relativePath "/foo/bar", extraLengthCutoff must be 0, to avoid a bad result.
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

func isSymbolicLinkToDirectory(relativePath, absoluteRootPath string, filesystemImpl afero.Fs) (bool, error) {
	linkReader, ok := filesystemImpl.(afero.LinkReader)
	if !ok {
		return false, nil
	}

	absolutePath := filepath.Join(absoluteRootPath, relativePath)
	linkTarget, err := linkReader.ReadlinkIfPossible(absolutePath)
	if err != nil {
		debug.PrintStack()
		return false, err
	}

	if !filepath.IsAbs(linkTarget) {
		linkTarget = filepath.Join(absoluteRootPath, filepath.Dir(relativePath), linkTarget)
	}

	stat, err := filesystemImpl.Stat(linkTarget)
	if err != nil {
		debug.PrintStack()
		return false, err
	}
	return stat.IsDir(), nil
}

// ScanDirectory returns the pointer to a (hierarchically-nested) Directory that is constructed from recursively walking
// the directory located at absoluteRootPath.
func ScanDirectory(absoluteRootPath string, filesystemImpl afero.Fs) (*Directory, error) {
	// Handle a special case that happens only during unit testing (where root is '\' when executed on Windows)
	if absoluteRootPath != "\\" {
		absRootPath, err := filepath.Abs(absoluteRootPath)
		if err != nil {
			debug.PrintStack()
			return nil, err
		}
		absoluteRootPath = absRootPath
	}

	directory := newDirectory(false)
	err := afero.Walk(filesystemImpl, absoluteRootPath, func(relativePath string, info fs.FileInfo, err error) error {
		if err != nil {
			debug.PrintStack()
			return err
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
				// We cannot trust the IsDir() output - it counts symbolic links to DIRS to be files
				isDir, err := isSymbolicLinkToDirectory(relativePath, absoluteRootPath, filesystemImpl)
				if err != nil {
					return err
				}
				if isDir {
					fileType = TypeDirSymlink
				}
			}

			err := directory.Add(relativePath, relativePath, absoluteRootPath, fileType, filesystemImpl)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		debug.PrintStack()
		return nil, err
	}

	return directory, nil
}
