package directory_checksum

import (
	"errors"
	"github.com/spf13/afero"
	"io/fs"
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

// ScanDirectory returns the pointer to a (hierarchically-nested) Directory that is constructed from recursively walking
// the directory located at absoluteRootPath.
func ScanDirectory(absoluteRootPath string, filesystemImpl afero.Fs, osWrapper OsWrapper) (*Directory, error) {
	absoluteRootPath = filepath.FromSlash(absoluteRootPath)
	if absoluteRootPath == "." {
		absRoot, err := osWrapper.Getwd()
		if err != nil {
			return nil, err
		}
		absoluteRootPath = absRoot
	}

	directory := newDirectory()
	err := afero.Walk(filesystemImpl, absoluteRootPath, func(relativePath string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Walk() is happy to walk a FILE (instead of a dir) -> we have to manually check that a dir path was provided
		if absoluteRootPath == relativePath && !info.IsDir() {
			return errors.New("provided root path must point to a directory")
		}

		if relativePath != absoluteRootPath {
			relativePath = bendRelativePath(relativePath, absoluteRootPath)
			err := directory.Add(relativePath, relativePath, absoluteRootPath, info.IsDir(), filesystemImpl)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return directory, nil
}
