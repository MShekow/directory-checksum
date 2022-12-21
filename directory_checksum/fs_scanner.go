package directory_checksum

import (
	"io/fs"
	"log"
	"os"
)

// ScanDirectory returns the pointer to a (hierarchically-nested) Directory that is constructed from recursively walking
// the directory located at absoluteRootPath.
func ScanDirectory(absoluteRootPath string) *Directory {
	directory := newDirectory()
	fileSystem := os.DirFS(absoluteRootPath)
	err := fs.WalkDir(fileSystem, ".", func(relativePath string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Fatal(err)
		}
		if relativePath != "." {
			directory.Add(relativePath, relativePath, absoluteRootPath, d.IsDir())
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	return directory
}
