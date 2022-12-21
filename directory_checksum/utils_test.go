// Contains various utils needed only by testing code
package directory_checksum

import (
	"errors"
	"github.com/spf13/afero"
	"os"
	"time"
)

// Based on https://github.com/spf13/afero/issues/213

// fsWrapper wraps an afero.Fs so that we can mock the Fs.Open() method and the File.Read() method to raise an error
// while reading from a file.
type fsWrapper struct {
	Fs afero.Fs
}

func (w *fsWrapper) Open(name string) (afero.File, error) {
	file, err := w.Fs.Open(name)
	if err != nil {
		return nil, err
	}
	return &fileWrapper{file}, nil
}

type fileWrapper struct {
	afero.File
}

func (f fileWrapper) Read(_ []byte) (n int, err error) {
	err = errors.New("reading failed")
	return
}

// Note: all methods below are simply pass-through implementations

func (w *fsWrapper) Create(name string) (afero.File, error) {
	return w.Fs.Create(name)
}

func (w *fsWrapper) Mkdir(name string, perm os.FileMode) error {
	return w.Fs.Mkdir(name, perm)
}

func (w *fsWrapper) MkdirAll(path string, perm os.FileMode) error {
	return w.Fs.MkdirAll(path, perm)
}

func (w *fsWrapper) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	return w.Fs.OpenFile(name, flag, perm)
}

func (w *fsWrapper) Remove(name string) error {
	return w.Fs.Remove(name)
}

func (w *fsWrapper) RemoveAll(path string) error {
	return w.Fs.RemoveAll(path)
}

func (w *fsWrapper) Rename(oldname, newname string) error {
	return w.Fs.Rename(oldname, newname)
}

func (w *fsWrapper) Stat(name string) (os.FileInfo, error) {
	return w.Fs.Stat(name)
}

func (w *fsWrapper) Name() string {
	return w.Fs.Name()
}

func (w *fsWrapper) Chmod(name string, mode os.FileMode) error {
	return w.Fs.Chmod(name, mode)
}

func (w *fsWrapper) Chown(name string, uid, gid int) error {
	return w.Fs.Chown(name, uid, gid)
}

func (w *fsWrapper) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return w.Fs.Chtimes(name, atime, mtime)
}
